package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
	"time"
)

const (
	UserKey       = "user"
	ImprintKey    = "imprint"
	AuthCookieKey = "wakapi_auth"
)

type MigrationFunc func(db *gorm.DB) error

type KeyStringValue struct {
	Key   string `gorm:"primary_key"`
	Value string `gorm:"type:text"`
}

func (j *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Split(strings.Trim(string(b), "\""), ".")[0]
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	t := time.Unix(i, 0)
	*j = CustomTime(t)
	return nil
}

func (j *CustomTime) Scan(value interface{}) error {
	switch value.(type) {
	case string:
		t, err := time.Parse("2006-01-02 15:04:05-07:00", value.(string))
		if err != nil {
			return errors.New(fmt.Sprintf("unsupported date time format: %s", value))
		}
		*j = CustomTime(t)
	case int64:
		*j = CustomTime(time.Unix(value.(int64), 0))
		break
	case time.Time:
		*j = CustomTime(value.(time.Time))
		break
	default:
		return errors.New(fmt.Sprintf("unsupported type: %T", value))
	}
	return nil
}

func (j CustomTime) Value() (driver.Value, error) {
	return time.Time(j), nil
}

func (j CustomTime) String() string {
	t := time.Time(j)
	return t.Format("2006-01-02 15:04:05")
}

func (j CustomTime) Time() time.Time {
	return time.Time(j)
}
