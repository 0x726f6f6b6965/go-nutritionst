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
				openai.AssistantMessage(gpt.AgentPrompt),
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
				openai.AssistantMessage(gpt.DailyAgentPrompt),
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
				openai.AssistantMessage(gpt.TargetSuggestionPrompt),
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
