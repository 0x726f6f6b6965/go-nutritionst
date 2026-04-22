package v2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const agentPrompt = `
You are a professional nutrition analyst.

Input:
- image (user meal photo)
- CONTEXT (daily nutrition status)
- gender
- age
- height (cm)
- current_weight (kg)
- target_weight (kg)
- timeframe (months)


Task:
1. Determine whether the image contains food.
2. If yes:
   - Estimate meal content and portion size based on visual cues
   - Aggregate into ONE dishes entry
   - Provide total estimated nutrition
3. Based on image + CONTEXT + target_weight:
   - Generate ONE integrated description including:
     (a) meal analysis
     (b) daily summary
     (c) next meal adjustment

Food Detection Rules:
- If not food:
  - set "is_food": false
  - set "dishes": []
  - explain "image_quality_issues"
- If partially visible / uncertain:
  - still "is_food": true
  - lower "foodness_confidence"
  - explain in "warnings"

Estimation Rules:
- Use visual references (utensils, hand size, container, packaging)
- If uncertain → use conservative reasonable estimate and explain in "notes"

Next Meal Adjustment Rules:
- Must include protein / carbs / fat
- Each must include:
  - direction (increase or decrease)
  - delta_g
  - delta_pct

General Constraints:
- All values must be >= 0
- Do NOT output NaN / Infinity
- Keep estimation realistic and consistent

Tone:
- Positive, supportive, professional
- Like real nutrition expert bestie 
- Include emoji

Output:
- Traditional Chinese
- description must be a SINGLE paragraph combining:
  meal analysis + daily summary + next meal adjustment
`

const targetSuggestionPrompt = `
You are a professional nutrition coach. Mentioned user target weight and timeframe.
Check the user's target weight and timeframe is reasonable or too aggressive,
based on the National Institute of Health guidelines and general medical consensus.

Input:
- gender
- age
- height (cm)
- current_weight (kg)
- target_weight (kg)
- timeframe (months)

Task:
- Based on user profile and target weight, suggest:
  1. A realistic target weight
  2. A timeframe (months)

Rules:
target is condisered with target_weight and target_timeframe combination, whether gain or lose weight:
- If healthy + reasonable → encourage
- If aggressive → warn about pace
- If underweight → warn about health risk
- If both → strongly advise adjustment
  
Tone:
- Positive, supportive, professional
- Like real nutrition expert bestie 
- Include emoji


Output:
- 只能是繁體中文 (Golden rule)
- 80-150 characters
- Must include:
  - Actionable suggestion (Target weight and timeframe adjustment if needed)
  - Clear judgment (reasonable / too aggressive)
  - Help you with the nutrition and diet plan adjustments to achieve the target safely.
  
Hard Constraints:
- Must output Traditional Chinese
- Double check the every character in the output is 繁體中文
`

const dailyAgentPrompt = `
You are a professional nutrition analyst.

Input:
- meals_today (list with ai_reply and totals)
- water_intake_ml
- CONTEXT
- gender
- age
- height (cm)
- current_weight (kg)
- target_weight (kg)
- timeframe (months)

Task:
Generate a "daily nutrition summary + coaching suggestion".

Tone:
- Positive, supportive, professional
- Like real nutrition expert bestie 
- May include emoji


Aggregation
- Sum all meal totals → day_totals


Target Calculation

BMR (Mifflin-St Jeor):
- male: 10W + 6.25H - 5A + 5
- female: 10W + 6.25H - 5A - 161

TDEE = BMR * activity_factor
(default activity_factor = 1.4)

Calories target:
- fat_loss: TDEE - 300~500 (max deficit 700)
- maintenance: TDEE
- muscle_gain: TDEE + 200~300

Protein:
- base: 1.6 g/kg
- fat_loss: 1.8-2.2 g/kg

Carbs/Fat:
- protein kcal = protein_g * 4
- remaining kcal:
  carbs 40-50%, fat 20-30%
- fat ≥ 0.6 g/kg

Water:
- 30-40 ml/kg
- if high protein → +10%


3. Fallback (if missing profile)
- calories: 1900 kcal
- protein: 80 g
- sodium: 2000 mg (limit 2300 mg)


4. Compliance
No extra description, just compliance status for each nutrient based on target and actual intake.

calories:
- ±10% → "完美"
- > → "高"
- < → "低"

protein:
- <90% → "不足"
- 90-140% → "適量"
- >140% → "過量"

sodium:
- ≤2300 → "正常"
- >2300 → "高"

deltas = actual - target


5. Insights
- 2-5 key observations
- based on meals + distribution


6. Coaching
- 3-5 actionable suggestions
- practical and daily-life friendly


7. Data Quality
- complete → []
- else → list issues

8. Constraints
- no NaN / Infinity
- no meaningless negative values

Output:
- description = ONE paragraph including:
  daily aggregation + compliance + insights + coaching
- All content must in Traditional Chinese
- No extra text
`

type Client struct {
	client              openai.Client
	model               string
	mealResponseSchema  openai.ResponseFormatJSONSchemaJSONSchemaParam
	dailyResponseSchema openai.ResponseFormatJSONSchemaJSONSchemaParam
}

func GenerateSchema[T any]() any {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

// Generate the JSON schema at initialization time
var (
	AIMealResponseSchema  = GenerateSchema[gpt.AIMealResponse]()
	AIDailyResponseSchema = GenerateSchema[gpt.AIDailyResponse]()
)

func NewClient(apiKey string) *Client {
	mealResponseSchema := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "meal_response",
		Description: openai.String("Notable information about a meal"),
		Schema:      AIMealResponseSchema,
		Strict:      openai.Bool(true),
	}
	dailyResponseSchema := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "daily_response",
		Description: openai.String("Notable information about a day"),
		Schema:      AIDailyResponseSchema,
		Strict:      openai.Bool(true),
	}
	return &Client{
		client:              openai.NewClient(option.WithAPIKey(apiKey)),
		model:               openai.ChatModelGPT4_1Mini,
		mealResponseSchema:  mealResponseSchema,
		dailyResponseSchema: dailyResponseSchema,
	}
}

func (c *Client) GetMealInfo(ctx context.Context, mealInfo *gpt.MealInfoWithImage) (*gpt.AIMealResponse, int64, error) {
	imgBase64 := base64.StdEncoding.EncodeToString(mealInfo.Image)
	imgURL := fmt.Sprintf("data:image/jpeg;base64,%s", imgBase64)
	fn := func(nctx context.Context) (res *openai.ChatCompletion, err error) {
		return c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.AssistantMessage(agentPrompt),
				openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
					openai.TextContentPart(fmt.Sprintf("這是%s, 請分析這張照片（繁體中文輸出）。\n%s", mealInfo.Description, mealInfo.UserProfile)),
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{URL: imgURL}),
				}),
			},
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: c.mealResponseSchema},
			},
			// Only certain models can perform structured outputs
			Model:               c.model,
			MaxCompletionTokens: openai.Int(1200),
		})
	}

	chat, err := Retry(ctx, 3, fn)
	var usage int64
	if chat != nil {
		usage = chat.Usage.TotalTokens
	}
	if err != nil {
		return nil, usage, err
	}
	var mealResponse gpt.AIMealResponse
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &mealResponse)
	if err != nil {
		return nil, usage, err
	}
	return &mealResponse, usage, nil
}

func (c *Client) GetMealDailyInfo(ctx context.Context, dailyInfo *gpt.DailyInfo) (*gpt.AIDailyResponse, int64, error) {
	jsonDailyInfo, err := json.Marshal(dailyInfo)
	if err != nil {
		return nil, 0, err
	}
	fn := func(nctx context.Context) (res *openai.ChatCompletion, err error) {
		return c.client.Chat.Completions.New(nctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.AssistantMessage(dailyAgentPrompt),
				openai.UserMessage(string(jsonDailyInfo)),
			},
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: c.dailyResponseSchema},
			},
			// Only certain models can perform structured outputs
			Model:               c.model,
			MaxCompletionTokens: openai.Int(1200),
		})
	}

	chat, err := Retry(ctx, 3, fn)
	var usage int64
	if chat != nil {
		usage = chat.Usage.TotalTokens
	}
	if err != nil {
		return nil, usage, err
	}
	var dailyResponse gpt.AIDailyResponse
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &dailyResponse)
	if err != nil {
		return nil, usage, err
	}
	return &dailyResponse, usage, nil
}

func (c *Client) GetTargetSuggestion(ctx context.Context, basicInfo *gpt.BasicUserInfo) (string, int64, error) {

	fn := func(nctx context.Context) (res *openai.ChatCompletion, err error) {
		return c.client.Chat.Completions.New(nctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.AssistantMessage(targetSuggestionPrompt),
				openai.UserMessage(basicInfo.UserProfile),
			},
			// Only certain models can perform structured outputs
			Model:               c.model,
			MaxCompletionTokens: openai.Int(1200),
		})
	}

	chat, err := Retry(ctx, 3, fn)
	var usage int64
	if chat != nil {
		usage = chat.Usage.TotalTokens
	}
	if err != nil {
		return "", usage, err
	}

	return chat.Choices[0].Message.Content, usage, nil
}

func Retry[T any](ctx context.Context, retryLimit int, fn func(ctx context.Context) (*T, error)) (*T, error) {
	nctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	duration := 10 * time.Second
	var (
		err    error
		result *T
	)
	for range retryLimit {
		result, err = fn(nctx)
		if err == nil {
			return result, nil
		}
		<-time.After(duration)
		duration += duration
	}
	return nil, fmt.Errorf("failed to retry after %d attempts, err: %w", retryLimit, err)
}
