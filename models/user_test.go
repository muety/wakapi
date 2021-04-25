package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUser_TZ(t *testing.T) {
	sut1, sut2 := &User{Location: ""}, &User{Location: "America/Los_Angeles"}
	pst, _ := time.LoadLocation("America/Los_Angeles")
	_, offset := time.Now().Zone()

	assert.Equal(t, time.Local, sut1.TZ())
	assert.Equal(t, pst, sut2.TZ())

	assert.InDelta(t, time.Duration(offset*int(time.Second)), sut1.TZOffset(), float64(1*time.Second))
	assert.InDelta(t, time.Duration(-7*int(time.Hour)), sut2.TZOffset(), float64(1*time.Second))
}
