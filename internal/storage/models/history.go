package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/Masterminds/squirrel"
)

type Meal int

const (
	MealUnknown   Meal = 0
	MealBreakfast Meal = 1
	MealLunch     Meal = 2
	MealDinner    Meal = 3
	MealSnack     Meal = 4
)

type MealHistory struct {
	ID            int       `json:"id" db:"id"`
	RequestID     string    `json:"request_id" db:"request_id"`
	LineID        string    `json:"line_id" db:"line_id"`
	Meal          Meal      `json:"meal" db:"meal"`
	Description   string    `json:"description" db:"description"`
	Photo         []byte    `json:"photo" db:"photo"`
	CaloriesKcal  float64   `json:"calories_kcal" db:"calories_kcal"`
	ProteinG      float64   `json:"protein_g" db:"protein_g"`
	CarbsG        float64   `json:"carbs_g" db:"carbs_g"`
	FatG          float64   `json:"fat_g" db:"fat_g"`
	SodiumMg      float64   `json:"sodium_mg" db:"sodium_mg"`
	AIDescription string    `json:"ai_description" db:"ai_description"`
	AIWarnings    string    `json:"ai_warnings" db:"ai_warnings"`
	AISuggest     string    `json:"ai_suggest" db:"ai_suggest"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type MealDaily struct {
	ID                           int       `json:"id" db:"id"`
	Date                         time.Time `json:"date" db:"date"`
	RequestID                    string    `json:"request_id" db:"request_id"`
	LineID                       string    `json:"line_id" db:"line_id"`
	BreakfastMeals               int       `json:"breakfast_meals" db:"breakfast_meals"`
	LunchMeals                   int       `json:"lunch_meals" db:"lunch_meals"`
	DinnerMeals                  int       `json:"dinner_meals" db:"dinner_meals"`
	SnackMeals                   int       `json:"snack_meals" db:"snack_meals"`
	TotalCaloriesKcal            float64   `json:"total_calories_kcal" db:"total_calories_kcal"`
	TotalProteinG                float64   `json:"total_protein_g" db:"total_protein_g"`
	TotalCarbsG                  float64   `json:"total_carbs_g" db:"total_carbs_g"`
	TotalFatG                    float64   `json:"total_fat_g" db:"total_fat_g"`
	TotalSodiumMg                float64   `json:"total_sodium_mg" db:"total_sodium_mg"`
	ComplianceCalories           string    `json:"compliance_calories" db:"compliance_calories"`
	ComplianceProtein            string    `json:"compliance_protein" db:"compliance_protein"`
	ComplianceSodium             string    `json:"compliance_sodium" db:"compliance_sodium"`
	ComplianceDeltasCaloriesKcal float64   `json:"compliance_deltas_calories_kcal" db:"compliance_deltas_calories_kcal"`
	ComplianceDeltasProteinG     float64   `json:"compliance_deltas_protein_g" db:"compliance_deltas_protein_g"`
	ComplianceDeltasSodiumMg     float64   `json:"compliance_deltas_sodium_mg" db:"compliance_deltas_sodium_mg"`
	Insights                     []string  `json:"insights" db:"insights"`
	TodayCoaching                []string  `json:"today_coaching" db:"today_coaching"`
	DataQualityIssues            []string  `json:"data_quality_issues" db:"data_quality_issues"`
	CreatedAt                    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at" db:"updated_at"`
}

func GetMealsDescription(meals *MealDaily) string {
	snacks := meals.SnackMeals
	breakfast := meals.BreakfastMeals
	lunch := meals.LunchMeals
	dinner := meals.DinnerMeals
	description := []string{}
	if breakfast > 0 {
		description = append(description, fmt.Sprintf("%d份早餐", breakfast))
	}
	if lunch > 0 {
		description = append(description, fmt.Sprintf("%d份午餐", lunch))
	}
	if dinner > 0 {
		description = append(description, fmt.Sprintf("%d份晚餐", dinner))
	}
	if snacks > 0 {
		description = append(description, fmt.Sprintf("%d份點心", snacks))
	}
	return strings.Join(description, "、")
}

func GetMealsIntFromMeals(meals []Meal, q *query.Query) int {
	var (
		breakfastMeals int
		lunchMeals     int
		dinnerMeals    int
		snackMeals     int
	)
	for _, meal := range meals {
		switch meal {
		case MealBreakfast:
			breakfastMeals += 1
		case MealLunch:
			lunchMeals += 1
		case MealDinner:
			dinnerMeals += 1
		case MealSnack:
			snackMeals += 1
		}
	}
	q.AddFilter(squirrel.Eq{"breakfast_meals": breakfastMeals}).
		AddFilter(squirrel.Eq{"lunch_meals": lunchMeals}).
		AddFilter(squirrel.Eq{"dinner_meals": dinnerMeals}).
		AddFilter(squirrel.Eq{"snack_meals": snackMeals})
	return breakfastMeals + lunchMeals + dinnerMeals + snackMeals
}
