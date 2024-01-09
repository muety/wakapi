package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	UserKey               = "user"
	ImprintKey            = "imprint"
	AuthCookieKey         = "wakapi_auth"
	PersistentIntervalKey = "wakapi_summary_interval"
)

type KeyStringValue struct {
	Key   string `gorm:"primary_key"`
	Value string `gorm:"type:text"`
}

type Interval struct {
	Start time.Time
	End   time.Time
}

type KeyedInterval struct {
	Interval
	Key *IntervalKey
}

// CustomTime is a wrapper type around time.Time, mainly used for the purpose of transparently unmarshalling Python timestamps in the format <sec>.<nsec> (e.g. 1619335137.3324468)
type CustomTime time.Time

// sql server doesn't allow multiple columns with timestamp type in a column
// So we need to change the type to datetimeoffset
func (j CustomTime) GormDBDataType(db *gorm.DB, field *schema.Field) string {

	t := "timestamp"

	if db.Config.Dialector.Name() == (sqlserver.Dialector{}).Name() {
		t = "datetimeoffset"
	}

	if scale, ok := field.TagSettings["TIMESCALE"]; ok {
		t += fmt.Sprintf("(%s)", scale)
	}

	return t
}

func (j *CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.T())
}

func (j *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	ts, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	t := time.Unix(0, int64(ts*1e9)) // ms to ns
	*j = CustomTime(t)
	return nil
}

func (j *CustomTime) Scan(value interface{}) error {
	var (
		t   time.Time
		err error
	)

	switch value.(type) {
	case string:
		// with sqlite, some queries (like GetLastByUser()) return dates as strings,
		// however, most of the time they are returned as time.Time
		t, err = time.Parse("2006-01-02 15:04:05-07:00", value.(string))
		if err != nil {
			return errors.New(fmt.Sprintf("unsupported date time format: %s", value))
		}
	case time.Time:
		t = value.(time.Time)
		break
	default:
		return errors.New(fmt.Sprintf("unsupported type: %T", value))
	}

	t = time.Unix(0, (t.UnixNano()/int64(time.Millisecond))*int64(time.Millisecond)) // round to millisecond precision
	*j = CustomTime(t)

	return nil
}

func (j CustomTime) Value() (driver.Value, error) {
	t := time.Unix(0, j.T().UnixNano()/int64(time.Millisecond)*int64(time.Millisecond)) // round to millisecond precision
	return t, nil
}

func (j *CustomTime) Hash() (uint64, error) {
	return uint64((j.T().UnixNano() / 1000) / 1000), nil
}

func (j CustomTime) String() string {
	return j.T().String()
}

func (j CustomTime) T() time.Time {
	return time.Time(j)
}

func (j CustomTime) Valid() bool {
	return j.T().Unix() >= 0
}
