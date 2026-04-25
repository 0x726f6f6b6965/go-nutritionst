package v3

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
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	client                 openai.Client
	model                  string
	mealResponseSchema     responses.ResponseFormatTextConfigUnionParam
	dailyResponseSchema    responses.ResponseFormatTextConfigUnionParam
	targetSuggestionSchema responses.ResponseFormatTextConfigUnionParam
}

func GenerateSchema[T any]() map[string]any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	data, _ := json.Marshal(schema)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}

// Generate the JSON schema at initialization time
var (
	AIMealResponseSchema     = GenerateSchema[gpt.AIMealResponse]()
	AIDailyResponseSchema    = GenerateSchema[gpt.AIDailyResponse]()
	AITargetSuggestionSchema = GenerateSchema[gpt.AITargetSuggestionResponse]()
)

func NewClient(apiKey string) *Client {
	return &Client{
		client:                 openai.NewClient(option.WithAPIKey(apiKey)),
		model:                  openai.ChatModelGPT4_1Mini,
		mealResponseSchema:     responses.ResponseFormatTextConfigParamOfJSONSchema("meal_response", AIMealResponseSchema),
		dailyResponseSchema:    responses.ResponseFormatTextConfigParamOfJSONSchema("daily_response", AIDailyResponseSchema),
		targetSuggestionSchema: responses.ResponseFormatTextConfigParamOfJSONSchema("target_suggestion_response", AITargetSuggestionSchema),
	}
}

func (c *Client) GetMealInfo(ctx context.Context, mealInfo *gpt.MealInfoWithImage) (*gpt.AIMealResponse, int64, error) {
	imgBase64 := base64.StdEncoding.EncodeToString(mealInfo.Image)
	imgURL := fmt.Sprintf("data:image/jpeg;base64,%s", imgBase64)
	fn := func(nctx context.Context) (*responses.Response, error) {
		req := responses.ResponseNewParams{
			Model: c.model,
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: []responses.ResponseInputItemUnionParam{
					{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Role: "system",
							Content: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Text: gpt.AgentPrompt,
									},
								},
							},
						},
					},
					{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Role: "user",
							Content: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Text: fmt.Sprintf("這是%s, 請分析這張照片（繁體中文輸出）。\n%s", mealInfo.Description, mealInfo.UserProfile),
									},
								},
								responses.ResponseInputContentUnionParam{
									OfInputImage: &responses.ResponseInputImageParam{
										ImageURL: openai.Opt(imgURL),
									},
								},
							},
						},
					},
				},
			},
			Text: responses.ResponseTextConfigParam{
				Format: c.mealResponseSchema,
			},
			Store:           openai.Bool(true),
			MaxOutputTokens: openai.Int(1200),
		}
		return c.client.Responses.New(ctx, req)
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
	err = json.Unmarshal([]byte(chat.OutputText()), &mealResponse)
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
	fn := func(nctx context.Context) (*responses.Response, error) {
		req := responses.ResponseNewParams{
			Model: c.model,
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: []responses.ResponseInputItemUnionParam{
					{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Role: "system",
							Content: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Text: gpt.DailyAgentPrompt,
									},
								},
							},
						},
					},
					{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Role: "user",
							Content: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Text: string(jsonDailyInfo),
									},
								},
							},
						},
					},
				},
			},
			Text: responses.ResponseTextConfigParam{
				Format: c.dailyResponseSchema,
			},
			Store:           openai.Bool(true),
			MaxOutputTokens: openai.Int(1200),
		}
		return c.client.Responses.New(nctx, req)
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
	err = json.Unmarshal([]byte(chat.OutputText()), &dailyResponse)
	if err != nil {
		return nil, usage, err
	}
	return &dailyResponse, usage, nil
}

func (c *Client) GetTargetSuggestion(ctx context.Context, basicInfo *gpt.BasicUserInfo) (string, int64, error) {

	fn := func(nctx context.Context) (*responses.Response, error) {
		req := responses.ResponseNewParams{
			Model: c.model,
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: []responses.ResponseInputItemUnionParam{
					{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Role: "system",
							Content: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Text: gpt.TargetSuggestionPrompt,
									},
								},
							},
						},
					},
					{
						OfInputMessage: &responses.ResponseInputItemMessageParam{
							Role: "user",
							Content: responses.ResponseInputMessageContentListParam{
								responses.ResponseInputContentUnionParam{
									OfInputText: &responses.ResponseInputTextParam{
										Text: basicInfo.UserProfile,
									},
								},
							},
						},
					},
				},
			},
			Text: responses.ResponseTextConfigParam{
				Format: c.targetSuggestionSchema,
			},
			Store:           openai.Bool(true),
			MaxOutputTokens: openai.Int(1200),
		}
		return c.client.Responses.New(nctx, req)
	}

	chat, err := Retry(ctx, 3, fn)
	var usage int64
	if chat != nil {
		usage = chat.Usage.TotalTokens
	}
	if err != nil {
		return "", usage, err
	}

	return chat.OutputText(), usage, nil
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
