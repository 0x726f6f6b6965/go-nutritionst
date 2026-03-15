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
	ID            int       `json:"id" db:"id"`
	LineID        string    `json:"line_id" db:"line_id"`
	Height        float64   `json:"height" db:"height"`
	Weight        float64   `json:"weight" db:"weight"`
	Age           int       `json:"age" db:"age"`
	Gender        Gender    `json:"gender" db:"gender"`
	MaxDailyToken int64     `json:"max_daily_token" db:"max_daily_token"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
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
	return fmt.Sprintf("User Profile: Gender: %s, Age: %d, Height: %.1fcm, Weight: %.1fkg", genderStr, u.Age, u.Height, u.Weight)
}
