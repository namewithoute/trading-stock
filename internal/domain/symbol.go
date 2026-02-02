package domain

import "time"

type Symbol struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
