package models

import "time"

// Well представляет данные о скважине
type Well struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Depth        float64   `json:"depth"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`
	Productivity float64   `json:"productivity"`
	DrillingDate time.Time `json:"drilling_date"`
	Field        string    `json:"field"`
	Operator     string    `json:"operator"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
