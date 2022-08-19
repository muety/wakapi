package models

// Support Wakapi and WakaTime range / interval identifiers
// See https://wakatime.com/developers/#summaries
var (
	IntervalToday              = &IntervalKey{"today", "Today"}
	IntervalYesterday          = &IntervalKey{"day", "yesterday", "Yesterday"}
	IntervalThisWeek           = &IntervalKey{"week", "This Week"}
	IntervalLastWeek           = &IntervalKey{"Last Week"}
	IntervalThisMonth          = &IntervalKey{"month", "This Month"}
	IntervalLastMonth          = &IntervalKey{"Last Month"}
	IntervalThisYear           = &IntervalKey{"year"}
	IntervalPast7Days          = &IntervalKey{"7_days", "last_7_days", "Last 7 Days"}
	IntervalPast7DaysYesterday = &IntervalKey{"Last 7 Days from Yesterday"}
	IntervalPast14Days         = &IntervalKey{"Last 14 Days"}
	IntervalPast30Days         = &IntervalKey{"30_days", "last_30_days", "Last 30 Days"}
	IntervalPast6Months		   = &IntervalKey{"6_months", "last_6_months"}
	IntervalPast12Months       = &IntervalKey{"12_months", "last_12_months", "last_year"}
	IntervalAny                = &IntervalKey{"any", "all_time"}
)

var AllIntervals = []*IntervalKey{
	IntervalToday,
	IntervalYesterday,
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
