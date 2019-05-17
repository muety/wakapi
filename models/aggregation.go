package models

import (
	"database/sql/driver"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	NAggregationTypes   uint8 = 4
	AggregationProject  uint8 = 0
	AggregationLanguage uint8 = 1
	AggregationEditor   uint8 = 2
	AggregationOS       uint8 = 3
)

type Aggregation struct {
	gorm.Model
	User     *User         `gorm:"not null; association_foreignkey:ID"`
	UserID   string        `gorm:"not null; index:idx_user,idx_type_time_user"`
	FromTime *time.Time    `gorm:"not null; index:idx_from,idx_type_time_user; default:now()"`
	ToTime   *time.Time    `gorm:"not null; index:idx_to,idx_type_time_user; default:now()"`
	Duration time.Duration `gorm:"-"`
	Type     uint8         `gorm:"not null; index:idx_type,idx_type_time_user"`
	Items    []AggregationItem
}

type AggregationItem struct {
	ID            uint   `gorm:"primary_key; auto_increment"`
	AggregationID uint   `gorm:"not null; association_foreignkey:ID"`
	Key           string `gorm:"not null"`
	Total         ScannableDuration
}

type ScannableDuration time.Duration

func (d *ScannableDuration) Scan(value interface{}) error {
	*d = ScannableDuration(*d) * ScannableDuration(time.Second)
	return nil
}

func (d ScannableDuration) Value() (driver.Value, error) {
	return int64(time.Duration(d) / time.Second), nil
}
