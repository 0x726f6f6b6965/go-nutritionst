package models

import "time"

type Usage struct {
	ID           int       `json:"id" db:"id"`
	LineID       string    `json:"line_id" db:"line_id"`
	Usage        int64     `json:"usage" db:"used_token"`
	LastUsedDate string    `json:"last_used_date" db:"last_used_date"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

func (u *Usage) GetLastUsedDate() (time.Time, error) {
	if u.LastUsedDate == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", u.LastUsedDate)
}
