package models

import (
	conf "github.com/muety/wakapi/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUser_TZ(t *testing.T) {
	sut1, sut2 := &User{Location: ""}, &User{Location: "America/Los_Angeles"}
	pst, _ := time.LoadLocation("America/Los_Angeles")
	_, offset1 := time.Now().Zone()
	_, offset2 := time.Now().In(pst).Zone()

	assert.Equal(t, time.Local, sut1.TZ())
	assert.Equal(t, pst, sut2.TZ())

	assert.InDelta(t, time.Duration(offset1*int(time.Second)), sut1.TZOffset(), float64(1*time.Second))
	assert.InDelta(t, time.Duration(offset2*int(time.Second)), sut2.TZOffset(), float64(1*time.Second))
}

func TestUser_MinDataAge(t *testing.T) {
	c := conf.Load("")

	var sut *User

	// test with unlimited retention time / clean-up disabled
	c.App.DataRetentionMonths = -1
	c.Subscriptions.Enabled = false
	sut = &User{}
	assert.Zero(t, sut.MinDataAge())

	// test with limited retention time / clean-up enabled, and subscriptions disabled
	c.App.DataRetentionMonths = 1
	c.Subscriptions.Enabled = false
	sut = &User{}
	assert.WithinRange(t, sut.MinDataAge(), time.Now().AddDate(0, -1, -1), time.Now().AddDate(0, -1, 1))

	// test with limited retention time, subscriptions enabled, and user hasn't got one
	c.App.DataRetentionMonths = 1
	c.Subscriptions.Enabled = true
	sut = &User{}
	assert.WithinRange(t, sut.MinDataAge(), time.Now().AddDate(0, -1, -1), time.Now().AddDate(0, -1, 1))

	// test with limited retention time, subscriptions disabled, but user still got (an expired) one
	c.App.DataRetentionMonths = 1
	c.Subscriptions.Enabled = false
	until2 := CustomTime(time.Now().AddDate(0, 0, -1))
	sut = &User{SubscribedUntil: &until2}
	assert.WithinRange(t, sut.MinDataAge(), time.Now().AddDate(0, -1, -1), time.Now().AddDate(0, -1, 1))

	// test with limited retention time, subscriptions enabled, and user has got one
	c.App.DataRetentionMonths = 1
	c.Subscriptions.Enabled = true
	until1 := CustomTime(time.Now().AddDate(0, 1, 0))
	sut = &User{SubscribedUntil: &until1}
	assert.Zero(t, sut.MinDataAge())
}
