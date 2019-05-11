package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	AggregationProject  int = 0
	AggregationLanguage int = 1
	AggregationEditor   int = 2
	AggregationOS       int = 3
)

type Aggregation struct {
	gorm.Model
	User     *User         `gorm:"not null; association_foreignkey:ID"`
	UserID   string        `gorm:"not null; index:idx_user,idx_type_time_user"`
	From     time.Time     `gorm:"not null; index:idx_from,idx_type_time_user; default:now()"`
	To       time.Time     `gorm:"not null; index:idx_to,idx_type_time_user; default:now()"`
	Duration time.Duration `gorm:"-"`
	Type     uint8         `gorm:"not null; index:idx_type,idx_type_time_user"`
	Items    []AggregationItem
}

type AggregationItem struct {
	AggregationID uint   `gorm:"not null; association_foreignkey:ID"`
	Key           string `gorm:"not null"`
	Total         time.Duration
}
