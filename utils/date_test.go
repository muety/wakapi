package utils

import (
	"github.com/muety/wakapi/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	tzLocal *time.Location
	tzUtc   *time.Location
	tzCet   *time.Location
	tzPst   *time.Location
)

func init() {
	tzLocal = time.Local
	tzUtc, _ = time.LoadLocation("UTC")
	tzCet, _ = time.LoadLocation("Europe/Berlin")
	tzPst, _ = time.LoadLocation("America/Los_Angeles")
}

func TestDate_Ceil(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			"02 Jan 06 15:04 MST",
			"03 Jan 06 00:00 MST",
		},
		{
			"03 Jan 06 00:00 MST",
			"03 Jan 06 00:00 MST",
		},
	}

	for _, test := range tests {
		inDate, _ := time.Parse(time.RFC822, test.in)
		outDate, _ := time.Parse(time.RFC822, test.out)
		out := CeilDate(inDate)
		assert.Equal(t, outDate, out)
	}
}

func TestDate_StartOfDay(t *testing.T) {
	d1, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzLocal)
	d2, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzUtc)
	d3, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzPst)
	d4, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzCet)

	t1, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 00:00:00", tzLocal)
	t2, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 00:00:00", tzUtc)
	t3, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 00:00:00", tzPst)
	t4, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 00:00:00", tzCet)

	assert.Equal(t, t1, StartOfDay(d1))
	assert.Equal(t, t2, StartOfDay(d2))
	assert.Equal(t, t3, StartOfDay(d3))
	assert.Equal(t, t4, StartOfDay(d4))

	assert.Equal(t, tzLocal, StartOfDay(d1).Location())
	assert.Equal(t, tzUtc, StartOfDay(d2).Location())
	assert.Equal(t, tzPst, StartOfDay(d3).Location())
	assert.Equal(t, tzCet, StartOfDay(d4).Location())
}

func TestDate_EndOfDay(t *testing.T) {
	d1, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzLocal)
	d2, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzUtc)
	d3, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzPst)
	d4, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzCet)

	t1, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-26 00:00:00", tzLocal)
	t2, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-26 00:00:00", tzUtc)
	t3, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-26 00:00:00", tzPst)
	t4, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-26 00:00:00", tzCet)

	assert.Equal(t, t1, EndOfDay(d1))
	assert.Equal(t, t2, EndOfDay(d2))
	assert.Equal(t, t3, EndOfDay(d3))
	assert.Equal(t, t4, EndOfDay(d4))

	assert.Equal(t, tzLocal, EndOfDay(d1).Location())
	assert.Equal(t, tzUtc, EndOfDay(d2).Location())
	assert.Equal(t, tzPst, EndOfDay(d3).Location())
	assert.Equal(t, tzCet, EndOfDay(d4).Location())
}

func TestDate_StartOfWeek(t *testing.T) {
	d1, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzLocal)
	d2, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzUtc)
	d3, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzPst)
	d4, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-25 20:25:00", tzCet)

	t1, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-19 00:00:00", tzLocal)
	t2, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-19 00:00:00", tzUtc)
	t3, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-19 00:00:00", tzPst)
	t4, _ := time.ParseInLocation(config.SimpleDateTimeFormat, "2021-04-19 00:00:00", tzCet)

	assert.Equal(t, t1, StartOfWeek(d1))
	assert.Equal(t, t2, StartOfWeek(d2))
	assert.Equal(t, t3, StartOfWeek(d3))
	assert.Equal(t, t4, StartOfWeek(d4))

	assert.Equal(t, tzLocal, StartOfWeek(d1).Location())
	assert.Equal(t, tzUtc, StartOfWeek(d2).Location())
	assert.Equal(t, tzPst, StartOfWeek(d3).Location())
	assert.Equal(t, tzCet, StartOfWeek(d4).Location())
}

func TestDate_SplitRangeByDays(t *testing.T) {
	df1, _ := time.Parse(config.SimpleDateTimeFormat, "2021-04-25 20:25:00")
	dt1, _ := time.Parse(config.SimpleDateTimeFormat, "2021-04-28 06:45:00")
	df2 := df1
	dt2 := CeilDate(df1)
	df3 := df1
	dt3 := df1.Add(10 * time.Second)
	df4 := df1
	dt4 := df4

	result1 := SplitRangeByDays(df1, dt1)
	result2 := SplitRangeByDays(df2, dt2)
	result3 := SplitRangeByDays(df3, dt3)
	result4 := SplitRangeByDays(df4, dt4)

	assert.Len(t, result1, 4)
	assert.Len(t, result1[0], 2)
	assert.Equal(t, result1[0][0], df1)
	assert.Equal(t, result1[3][1], dt1)
	assert.Equal(t, result1[1][0].Hour()+result1[1][0].Minute()+result1[1][0].Second(), 0)
	assert.Equal(t, result1[2][0].Hour()+result1[2][0].Minute()+result1[2][0].Second(), 0)
	assert.Equal(t, result1[3][0].Hour()+result1[3][0].Minute()+result1[3][0].Second(), 0)
	assert.Equal(t, result1[1][0], result1[0][1])
	assert.Equal(t, result1[2][0], result1[1][1])
	assert.Equal(t, result1[3][0], result1[2][1])

	assert.Len(t, result2, 1)
	assert.Equal(t, result2[0][0], df2)
	assert.Equal(t, result2[0][1], dt2)

	assert.Len(t, result3, 1)
	assert.Equal(t, result3[0][0], df3)
	assert.Equal(t, result3[0][1], dt3)

	assert.Len(t, result4, 0)
}
