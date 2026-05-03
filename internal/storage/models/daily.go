package models

import (
	"time"

	"github.com/google/uuid"
)

type DailyRecord struct {
	ID             int64     `db:"id"`
	RequestID      uuid.UUID `db:"request_id"`
	LineID         string    `db:"line_id"`
	Date           string    `db:"date"`
	TotalWaterMl   *float64  `db:"total_water_ml"`
	TotalSleepHour *float64  `db:"total_sleep_hour"`
	BreakfastMeals int       `db:"breakfast_meals"`
	LunchMeals     int       `db:"lunch_meals"`
	DinnerMeals    int       `db:"dinner_meals"`
	SnackMeals     int       `db:"snack_meals"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
