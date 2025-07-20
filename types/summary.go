package types

import (
	"errors"
	"time"

	"github.com/muety/wakapi/models"
)

// SummaryRequest represents a request for summary generation
type SummaryRequest struct {
	From      time.Time
	To        time.Time
	User      *models.User
	Filters   *models.Filters
	SkipCache bool
}

// NewSummaryRequest creates a new summary request with the required parameters
func NewSummaryRequest(from, to time.Time, user *models.User) *SummaryRequest {
	return &SummaryRequest{
		From: from,
		To:   to,
		User: user,
	}
}

// WithFilters adds filters to the summary request
func (r *SummaryRequest) WithFilters(filters *models.Filters) *SummaryRequest {
	r.Filters = filters
	return r
}

// WithoutCache disables caching for this request
func (r *SummaryRequest) WithoutCache() *SummaryRequest {
	r.SkipCache = true
	return r
}

// Validate ensures the request has valid parameters
func (r *SummaryRequest) Validate() error {
	if r.User == nil {
		return errors.New("user is required")
	}
	if r.From.IsZero() || r.To.IsZero() {
		return errors.New("from and to times are required")
	}
	if !r.To.After(r.From) {
		return errors.New("to time must be after from time")
	}
	return nil
}

// ProcessingOptions configures how a summary should be processed
type ProcessingOptions struct {
	ApplyAliases       bool
	ApplyProjectLabels bool
	UseCache           bool
}

// DefaultProcessingOptions returns the default processing configuration
func DefaultProcessingOptions() *ProcessingOptions {
	return &ProcessingOptions{
		ApplyAliases:       true,
		ApplyProjectLabels: true,
		UseCache:           true,
	}
}

// WithAliases enables alias resolution
func (o *ProcessingOptions) WithAliases() *ProcessingOptions {
	o.ApplyAliases = true
	return o
}

// WithoutAliases disables alias resolution
func (o *ProcessingOptions) WithoutAliases() *ProcessingOptions {
	o.ApplyAliases = false
	return o
}

// WithProjectLabels enables project label resolution
func (o *ProcessingOptions) WithProjectLabels() *ProcessingOptions {
	o.ApplyProjectLabels = true
	return o
}

// WithoutProjectLabels disables project label resolution
func (o *ProcessingOptions) WithoutProjectLabels() *ProcessingOptions {
	o.ApplyProjectLabels = false
	return o
}

// WithoutCache disables caching
func (o *ProcessingOptions) WithoutCache() *ProcessingOptions {
	o.UseCache = false
	return o
}