package models

import (
	"strconv"
	"strings"
	"time"
)

type HeartbeatReqTime time.Time

type Heartbeat struct {
	User     string           `json:"user"`
	Entity   string           `json:"entity"`
	Type     string           `json:"type"`
	Category string           `json:"category"`
	Project  string           `json:"project"`
	Branch   string           `json:"branch"`
	Language string           `json:"language"`
	IsWrite  bool             `json:"is_write"`
	Time     HeartbeatReqTime `json:"time"`
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

func (j HeartbeatReqTime) String() string {
	t := time.Time(j)
	return t.Format("2006-01-02 15:04:05")
}
