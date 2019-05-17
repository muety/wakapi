package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type HeartbeatReqTime time.Time

type Heartbeat struct {
	gorm.Model
	User            *User             `json:"user" gorm:"not null; association_foreignkey:ID"`
	UserID          string            `json:"-" gorm:"not null; index:idx_time_user"`
	Entity          string            `json:"entity" gorm:"not null"`
	Type            string            `json:"type"`
	Category        string            `json:"category"`
	Project         string            `json:"project" gorm:"index:idx_project"`
	Branch          string            `json:"branch"`
	Language        string            `json:"language" gorm:"not null"`
	IsWrite         bool              `json:"is_write"`
	Editor          string            `json:"editor" gorm:"not null"`
	OperatingSystem string            `json:"operating_system" gorm:"not null"`
	Time            *HeartbeatReqTime `json:"time" gorm:"type:timestamp; default:now(); index:idx_time,idx_time_user"`
}

func (h *Heartbeat) Valid() bool {
	return h.User != nil && h.UserID != "" && h.Entity != "" && h.Language != "" && h.Editor != "" && h.OperatingSystem != "" && h.Time != nil
}

func (j *HeartbeatReqTime) UnmarshalJSON(b []byte) error {
	s := strings.Split(strings.Trim(string(b), "\""), ".")[0]
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	t := time.Unix(i, 0)
	*j = HeartbeatReqTime(t)
	return nil
}

func (j *HeartbeatReqTime) Scan(value interface{}) error {
	switch value.(type) {
	case int64:
		*j = HeartbeatReqTime(time.Unix(123456, 0))
		break
	case time.Time:
		*j = HeartbeatReqTime(value.(time.Time))
		break
	default:
		return errors.New(fmt.Sprintf("Unsupported type"))
	}
	return nil
}

func (j HeartbeatReqTime) Value() (driver.Value, error) {
	return time.Time(j), nil
}

func (j HeartbeatReqTime) String() string {
	t := time.Time(j)
	return t.Format("2006-01-02 15:04:05")
}

func (j HeartbeatReqTime) Time() time.Time {
	return time.Time(j)
}
