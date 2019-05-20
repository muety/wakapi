package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type HeartbeatReqTime time.Time

type Heartbeat struct {
	ID              uint              `gorm:"primary_key"`
	User            *User             `json:"-" gorm:"not null; index:idx_time_user"`
	UserID          string            `json:"-" gorm:"not null; index:idx_time_user"`
	Entity          string            `json:"entity" gorm:"not null"`
	Type            string            `json:"type"`
	Category        string            `json:"category"`
	Project         string            `json:"project"`
	Branch          string            `json:"branch"`
	Language        string            `json:"language"`
	IsWrite         bool              `json:"is_write"`
	Editor          string            `json:"editor"`
	OperatingSystem string            `json:"operating_system"`
	Time            *HeartbeatReqTime `json:"time" gorm:"type:timestamp; default:now(); index:idx_time,idx_time_user"`
}

func (h *Heartbeat) Valid() bool {
	return h.User != nil && h.UserID != "" && h.Time != nil
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
