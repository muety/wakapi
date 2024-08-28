package models

import (
	"github.com/lib/pq" // first use-age of pg specific things. This won't work with other dialects
)

type Client struct {
	ID         string         `json:"id" gorm:"primary_key"`
	UserID     string         `json:"user_id"`
	Name       string         `json:"name"`
	Currency   string         `json:"currency"`
	HourlyRate float64        `json:"hourly_rate"`
	Projects   pq.StringArray `json:"projects" gorm:"type:text[]"`
	CreatedAt  CustomTime     `json:"created_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
	UpdatedAt  CustomTime     `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
}

type NewClient struct {
	UserID     string   `json:"user_id"`
	Name       string   `json:"name"`
	Currency   string   `json:"currency"`
	HourlyRate float64  `json:"hourly_rate"`
	Projects   []string `json:"projects"`
}

type ClientUpdate struct {
	Name       string         `json:"name"`
	Currency   string         `json:"currency"`
	HourlyRate float64        `json:"hourly_rate"`
	Projects   pq.StringArray `json:"projects" gorm:"type:text[]"`
}
