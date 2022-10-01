package models

import "time"

type LeaderboardItem struct {
	ID        uint          `json:"-" gorm:"primary_key; size:32"`
	User      *User         `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserID    string        `json:"user_id" gorm:"not null; index:idx_leaderboard_user"`
	Rank      uint          `json:"rank" gorm:"->"`
	Interval  string        `json:"interval" gorm:"not null; size:32; index:idx_leaderboard_combined"`
	By        *uint8        `json:"aggregated_by" gorm:"index:idx_leaderboard_combined"` // pointer because nullable
	Total     time.Duration `json:"total" gorm:"not null" swaggertype:"primitive,integer"`
	Key       *string       `json:"key" gorm:"size:255"` // pointer because nullable
	CreatedAt CustomTime    `gorm:"type:timestamp; default:CURRENT_TIMESTAMP" swaggertype:"string" format:"date" example:"2006-01-02 15:04:05.000"`
}
