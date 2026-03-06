package gpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const agentPrompt = `
你是個人營養分析師。針對使用者上傳的照片，以及一段 user_profile, 請先判斷是否為餐點/食物。
菜名或食物名與營養可以統整在 dishes 不需要多道菜分開給出, est_nutrition 也只需要給出總和。
請依 CONTEXT 內容與照片, 同步完成「本餐分析 + 今日總結 + 下一餐調整」。
下一餐調整須包含個營養素分析，增加或減少的數字與百分比。
只輸出 JSON, 不要夾雜解釋或額外文字。鍵名與型別需完全符合下列結構:

{
  "is_food": true | false,
  "foodness_confidence": 0.0~1.0,
  "dishes": [
    {
      "name": "菜名或食物名稱(繁體中文)",
      "confidence": 0.0~1.0,
      "portion": {"unit": "g|ml|piece", "quantity": number},
      "est_nutrition": {
        "calories_kcal": number,
        "protein_g": number,
        "carbs_g": number,
        "fat_g": number,
        "sodium_mg": number
      },
      "notes": "可觀察到的烹調方式/醬料/配料等(繁中)"
    }
  ],
  "warnings": ["例如：油炸、醬料偏多、含糖飲等(繁中)"],
  "image_quality_issues": ["例如：過暗、模糊、僅拍到局部等(繁中)"],
  "explanations_zh": "用繁體中文簡短說明你的判斷依據與估算方法",
  "nutrition_report": "根據 est_nutrition 格式給出總結的數值",
  "personalized_advice": "結合 user_profile 的性別/年齡/身高/體重與目標值，給出本餐的具體建議（繁中）"
}

規則：
- 若非餐點，請將 "is_food": false, "dishes" 回傳空陣列，並寫明 "image_quality_issues" 與原因。
- 若難以判斷或只有部分食物, is_food 仍可為 true, 但請降低 foodness_confidence 並在 warnings 裡說明不確定之處。
- 估份量請以視覺參考（餐具尺寸、手掌、罐裝標示等）推測；無法估就用最保守合理值並在 notes 標記。
- 請避免低於 0 的數值;NaN/Infinity 一律不要出現。
- 僅輸出 JSON(response_format 已強制 JSON)。
`

const dailyAgentPrompt = `
你是個人營養分析師。針對使用者今天的飲食資料(meals_today)與一段 user_profile, 請整理成「單日飲食總結報告」。
你會收到 CONTEXT, 內容包含 user_profile, meals_today, meta。
請根據 CONTEXT 內容完成「單日營養統整 + 合規判斷 + 重點觀察 + 今日建議」。
口吻比較輕鬆俏皮，可以使用一些 emoji。
只輸出 JSON, 不要夾雜解釋或額外文字。鍵名與型別需完全符合下列結構:

{
  "day_totals": {
    "calories_kcal": number,
    "protein_g": number,
    "carbs_g": number,
    "fat_g": number,
    "sodium_mg": number
  },
  "compliance": {
    "calories": "偏低|接近|偏高",
    "protein": "不足|合理|過高",
    "sodium": "可|偏高",
    "deltas": {
      "calories_kcal": number,
      "protein_g": number,
      "sodium_mg": number
    }
  },
  "insights": [
    "2~5 條今天的飲食觀察（繁中）"
  ],
  "today_coaching": [
    "3~5 條具體可執行建議（繁中）"
  ],
  "data_quality_issues": [
    "資料缺漏或異常說明（繁中）"
  ]
}

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
- 僅輸出 JSON(response_format 已強制 JSON)。
`

type NutritionAPI interface {
	GetMealInfo(ctx context.Context, image []byte, mealName string, userProfile string) (*AIMealResponse, error)
}

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
		model:  openai.GPT4o, // Using GPT-4o as it is the latest vision capable model, or fallback to gpt-4-turbo
	}
}

type EstNutrition struct {
	CaloriesKcal float64 `json:"calories_kcal"`
	ProteinG     float64 `json:"protein_g"`
	CarbsG       float64 `json:"carbs_g"`
	FatG         float64 `json:"fat_g"`
	SodiumMg     float64 `json:"sodium_mg"`
}

type Dish struct {
	Name         string       `json:"name"`
	EstNutrition EstNutrition `json:"est_nutrition"`
}

type AIMealResponse struct {
	IsFood             bool     `json:"is_food"`
	Dishes             []Dish   `json:"dishes"`
	ExplanationsZh     string   `json:"explanations_zh"`
	PersonalizedAdvice string   `json:"personalized_advice"`
	Warnings           []string `json:"warnings"`
}

type MealInfo struct {
	Meal                int    `json:"meal"`
	Description         string `json:"meal_name"`
	UserProfile         string `json:"-"`
	PreviousDescription string `json:"previous_description"`
	PreviousAdvice      string `json:"previous_advice"`
}
type MealInfoWithImage struct {
	Image []byte `json:"image"`
	MealInfo
}

type DailyInfo struct {
	UserProfile string     `json:"user_profile"`
	MealsToday  []MealInfo `json:"meals_today"`
	Meta        string     `json:"meta"`
}

type Compliance struct {
	Calories string `json:"calories"`
	Protein  string `json:"protein"`
	Sodium   string `json:"sodium"`
	Deltas   struct {
		CaloriesKcal float64 `json:"calories_kcal"`
		ProteinG     float64 `json:"protein_g"`
		SodiumMg     float64 `json:"sodium_mg"`
	} `json:"deltas"`
}

type AIDailyResponse struct {
	DayTotals         EstNutrition `json:"day_totals"`
	Compliance        Compliance   `json:"compliance"`
	Insights          []string     `json:"insights"`
	TodayCoaching     []string     `json:"today_coaching"`
	DataQualityIssues []string     `json:"data_quality_issues"`
}

func (c *Client) GetMealInfo(ctx context.Context, mealInfo *MealInfoWithImage) (*AIMealResponse, error) {
	imgBase64 := base64.StdEncoding.EncodeToString(mealInfo.Image)
	imgURL := fmt.Sprintf("data:image/jpeg;base64,%s", imgBase64)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: agentPrompt,
				},
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: fmt.Sprintf("這是%s, 請分析這張照片（繁體中文輸出）。\n%s", mealInfo.Description, mealInfo.UserProfile),
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: imgURL,
							},
						},
					},
				},
			},
			MaxTokens: 1200,
		},
	)

	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	// Simple cleanup if markdown code blocks are present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var aiResp AIMealResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w, content: %s", err, content)
	}

	return &aiResp, nil
}

func (c *Client) GetMealDailyInfo(ctx context.Context, dailyInfo *DailyInfo) (*AIDailyResponse, error) {
	jsonDailyInfo, err := json.Marshal(dailyInfo)
	if err != nil {
		return nil, err
	}
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: dailyAgentPrompt,
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: string(jsonDailyInfo),
					},
				},
			},
		},
		MaxTokens: 1200,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	// Simple cleanup if markdown code blocks are present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var aiResp AIDailyResponse
	if err := json.Unmarshal([]byte(content), &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w, content: %s", err, content)
	}

	return &aiResp, nil
}
