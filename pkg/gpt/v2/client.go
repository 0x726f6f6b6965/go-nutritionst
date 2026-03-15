package v2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const agentPrompt = `
你是個人營養分析師。針對使用者上傳的照片，以及一段 user_profile, 請先判斷是否為餐點/食物。
菜名或食物名與營養可以統整在 dishes 不需要多道菜分開給出, est_nutrition 也只需要給出總和。
請依 CONTEXT 內容與照片, 同步完成「本餐分析 + 今日總結 + 下一餐調整」。
下一餐調整須包含個營養素分析，增加或減少的數字與百分比。

規則：
- 若非餐點，請將 "is_food": false, "dishes" 回傳空陣列，並寫明 "image_quality_issues" 與原因。
- 若難以判斷或只有部分食物, is_food 仍可為 true, 但請降低 foodness_confidence 並在 warnings 裡說明不確定之處。
- 估份量請以視覺參考（餐具尺寸、手掌、罐裝標示等）推測；無法估就用最保守合理值並在 notes 標記。
- 請避免低於 0 的數值;NaN/Infinity 一律不要出現。
- 回覆請使用繁體中文。
`

const dailyAgentPrompt = `
你是個人營養分析師。針對使用者今天的飲食資料(meals_today)與一段 user_profile, 請整理成「單日飲食總結報告」。
你會收到 CONTEXT, 內容包含 user_profile, meals_today, meta。
請根據 CONTEXT 內容完成「單日營養統整 + 合規判斷 + 重點觀察 + 今日建議」。
口吻比較輕鬆俏皮，可以使用一些 emoji。
只輸出 JSON, 不要夾雜解釋或額外文字。鍵名與型別需完全符合下列結構:

規則：
- 請將 MEALS_TODAY 中每餐的 totals 加總後，填入 day_totals。
- compliance 請根據 user_profile 推估每日目標後判斷：
  - 熱量目標:Mifflin-St Jeor + 活動係數 1.4
  - 蛋白目標:1.6 g/kg(以體重)
  - 鈉上限:2300 mg
- 若 user_profile 資料不足，請使用保守估計：
  - 熱量預設 2000 kcal
  - 蛋白預設 80 g
  - 鈉上限固定 2300 mg
- compliance 判定方式：
  - calories:相對目標 ±10% 以內 =「接近」，高於 =「偏高」，低於 =「偏低」
  - protein:<90% =「不足」,90~140% =「合理」,>140% =「過高」
  - sodium:<=2300 =「可」,>2300 =「偏高」
  - deltas = 今日總計 - 每日目標（可為負數）
- insights 請綜合各餐 ai_reply 與整日營養分布，整理 2~5 點觀察。
- today_coaching 請給 3~5 條具體、可執行、貼近日常的建議（例如：醬汁分開、補蛋白、飲料改無糖）。
- data_quality_issues 若資料完整請回傳空陣列 []。
- 請避免低於 0 的數值;NaN/Infinity 一律不要出現。
- 回覆請使用繁體中文。
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
		model:               openai.ChatModelGPT4oMini,
		mealResponseSchema:  mealResponseSchema,
		dailyResponseSchema: dailyResponseSchema,
	}
}

func (c *Client) GetMealInfo(ctx context.Context, mealInfo *gpt.MealInfoWithImage) (*gpt.AIMealResponse, int64, error) {
	imgBase64 := base64.StdEncoding.EncodeToString(mealInfo.Image)
	imgURL := fmt.Sprintf("data:image/jpeg;base64,%s", imgBase64)
	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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

	chat, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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
