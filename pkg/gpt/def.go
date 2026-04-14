package gpt

import (
	"context"
)

type NutritionAPI interface {
	GetMealInfo(ctx context.Context, mealInfo *MealInfoWithImage) (*AIMealResponse, int64, error)
	GetMealDailyInfo(ctx context.Context, dailyInfo *DailyInfo) (*AIDailyResponse, int64, error)
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

type EstNutrition struct {
	CaloriesKcal float64 `json:"calories_kcal" jsonschema_description:"Total calories in kcal for this dish"`
	ProteinG     float64 `json:"protein_g" jsonschema_description:"Total protein in g for this dish"`
	CarbsG       float64 `json:"carbs_g" jsonschema_description:"Total carbs in g for this dish"`
	FatG         float64 `json:"fat_g" jsonschema_description:"Total fat in g for this dish"`
	SodiumMg     float64 `json:"sodium_mg" jsonschema_description:"Total sodium in mg for this dish"`
}

type Dish struct {
	Name         string       `json:"name" jsonschema_description:"Name of the dish"`
	EstNutrition EstNutrition `json:"est_nutrition" jsonschema_description:"Estimated nutrition for this dish"`
}
type Deltas struct {
	CaloriesKcal float64 `json:"calories_kcal" jsonschema_description:"Today's insufficient or excessive calories in kcal"`
	ProteinG     float64 `json:"protein_g" jsonschema_description:"Today's insufficient or excessive protein in g"`
	SodiumMg     float64 `json:"sodium_mg" jsonschema_description:"Today's insufficient or excessive sodium in mg"`
}

type Compliance struct {
	Calories string `json:"calories" jsonschema_description:"Was today's calorie intake insufficient, excessive, or sufficient in Traditional Chinese."`
	Protein  string `json:"protein" jsonschema_description:"Was today's protein intake insufficient, excessive, or sufficient in Traditional Chinese."`
	Sodium   string `json:"sodium" jsonschema_description:"Was today's sodium intake insufficient, excessive, or sufficient in Traditional Chinese."`
	Deltas   Deltas `json:"deltas" jsonschema_description:"Today's insufficient or excessive calories, protein, and sodium in kcal, g, and mg."`
}

type AIMealResponse struct {
	IsFood             bool     `json:"is_food" jsonschema_description:"Whether the image is a food item."`
	Dishes             []Dish   `json:"dishes" jsonschema_description:"List of dishes detected in the image in Traditional Chinese."`
	ExplanationsZh     string   `json:"explanations_zh" jsonschema_description:"Briefly explain the judgment basis and estimation method in Traditional Chinese."`
	PersonalizedAdvice string   `json:"personalized_advice" jsonschema_description:"Based on the user_profile's gender/age/height/weight and target values, provide specific suggestions for this meal in Traditional Chinese."`
	Warnings           []string `json:"warnings" jsonschema_description:"Warnings regarding this meal, such as fried food, excessive sauce, and sugary drinks in Traditional Chinese."`
}

type AIDailyResponse struct {
	DayTotals         EstNutrition `json:"day_totals" jsonschema_description:"Total calories, protein, carbs, fat, and sodium in kcal, g, and mg for today in Traditional Chinese."`
	Compliance        Compliance   `json:"compliance" jsonschema_description:"Today's compliance with calorie, protein, and sodium intake in Traditional Chinese."`
	Insights          []string     `json:"insights" jsonschema_description:"The summarize 2-5 observations based on the ai_reply data for each meal and the overall daily nutritional distribution in Traditional Chinese."`
	TodayCoaching     []string     `json:"today_coaching" jsonschema_description:"Provide 3-5 specific, actionable, and practical suggestions (e.g., separate sauces, supplement with protein, switch to sugar-free beverages) in Traditional Chinese."`
	DataQualityIssues []string     `json:"data_quality_issues" jsonschema_description:"Data quality issues for today's meal in Traditional Chinese."`
}
