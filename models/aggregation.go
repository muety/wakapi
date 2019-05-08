package models

import "time"

type AggregationType string

const (
	AggregationProject  AggregationType = "project"
	AggregationLanguage AggregationType = "language"
	AggregationEditor   AggregationType = "editor"
	AggregationOS       AggregationType = "os"
)

type Aggregation struct {
	From  time.Time
	To    time.Time
	Type  AggregationType
	Items []AggregationItem
}

type AggregationItem struct {
	Key   string
	Total time.Duration
}
