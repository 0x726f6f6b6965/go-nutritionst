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
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
