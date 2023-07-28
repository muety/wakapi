package models

import (
	"unicode"
)

// Support Wakapi and WakaTime range / interval identifiers
// See https://wakatime.com/developers/#summaries
var (
	IntervalToday              = &IntervalKey{"today", "Today"}
	IntervalYesterday          = &IntervalKey{"day", "yesterday", "Yesterday"}
	IntervalPastDay            = &IntervalKey{"24_hours", "last_24_hours", "last_day", "Last 24 Hours"} // non-official one
	IntervalThisWeek           = &IntervalKey{"week", "This Week"}
	IntervalLastWeek           = &IntervalKey{"last_week", "Last Week"}
	IntervalThisMonth          = &IntervalKey{"month", "This Month"}
	IntervalLastMonth          = &IntervalKey{"last_month", "Last Month"}
	IntervalThisYear           = &IntervalKey{"year", "This Year"}
	IntervalPast7Days          = &IntervalKey{"7_days", "last_7_days", "Last 7 Days"}
	IntervalPast7DaysYesterday = &IntervalKey{"Last 7 Days from Yesterday"}
	IntervalPast14Days         = &IntervalKey{"14_days", "last_14_days", "Last 14 Days"}
	IntervalPast30Days         = &IntervalKey{"30_days", "last_30_days", "Last 30 Days"}
	IntervalPast6Months        = &IntervalKey{"6_months", "last_6_months", "Last 6 Months"}
	IntervalPast12Months       = &IntervalKey{"12_months", "last_12_months", "last_year", "Last 12 Months"}
	IntervalAny                = &IntervalKey{"any", "all_time", "All Time"}
)

var AllIntervals = []*IntervalKey{
	IntervalToday,
	IntervalYesterday,
	IntervalPastDay,
	IntervalThisWeek,
	IntervalLastWeek,
	IntervalThisMonth,
	IntervalLastMonth,
	IntervalThisYear,
	IntervalPast7Days,
	IntervalPast7DaysYesterday,
	IntervalPast14Days,
	IntervalPast30Days,
	IntervalPast6Months,
	IntervalPast12Months,
	IntervalAny,
}

type IntervalKey []string

func (k *IntervalKey) HasAlias(s string) bool {
	for _, e := range *k {
		if e == s {
			return true
		}
	}
	return false
}

func (k *IntervalKey) GetHumanReadable() string {
	for _, s := range *k {
		if unicode.IsUpper(rune(s[0])) {
			return s
		}
	}
	return ""
}
