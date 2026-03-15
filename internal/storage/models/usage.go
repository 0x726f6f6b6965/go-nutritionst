package models

import "time"

type Usage struct {
	ID        int       `json:"id" db:"id"`
	LineID    string    `json:"line_id" db:"line_id"`
	Usage     int64     `json:"usage" db:"used_token"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
