package models

import (
	"fmt"
	"time"
)

type Gender int

const (
	GenderUnknown Gender = 0
	GenderMale    Gender = 1
	GenderFemale  Gender = 2
)

type User struct {
	ID               int       `json:"id" db:"id"`
	LineID           string    `json:"line_id" db:"line_id"`
	Name             string    `json:"name" db:"name"`
	Height           float64   `json:"height" db:"height"`
	Weight           float64   `json:"weight" db:"weight"`
	TargetWeight     float64   `json:"target_weight" db:"target_weight"`
	TargetTimeframe  int       `json:"target_timeframe" db:"target_timeframe"`
	Age              int       `json:"age" db:"age"`
	Gender           Gender    `json:"gender" db:"gender"`
	MaxDailyToken    int64     `json:"max_daily_token" db:"max_daily_token"`
	MorningMsgSent   bool      `json:"morning_msg_sent" db:"morning_msg_sent"`
	EveningMsgSent   bool      `json:"evening_msg_sent" db:"evening_msg_sent"`
	BreakfastMsgSent bool      `json:"breakfast_msg_sent" db:"breakfast_msg_sent"`
	LunchMsgSent     bool      `json:"lunch_msg_sent" db:"lunch_msg_sent"`
	DinnerMsgSent    bool      `json:"dinner_msg_sent" db:"dinner_msg_sent"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) ToProfileString() string {
	var genderStr string
	switch u.Gender {
	case GenderMale:
		genderStr = "Male"
	case GenderFemale:
		genderStr = "Female"
	default:
		genderStr = "Unknown"
	}
	return fmt.Sprintf("User Profile: Gender: %s, Age: %d, Height: %.1fcm, Weight: %.1fkg, target_weight: %.1fkg, target_timeframe: %d months", genderStr, u.Age, u.Height, u.Weight, u.TargetWeight, u.TargetTimeframe)
}
