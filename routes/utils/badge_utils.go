package utils

import (
	"errors"
	"github.com/muety/wakapi/helpers"
	"github.com/muety/wakapi/models"
	"regexp"
)

const (
	intervalPattern     = `interval:([a-z0-9_]+)`
	entityFilterPattern = `(project|os|editor|language|machine|label):([^:?&/]+)`
)

var (
	intervalReg     *regexp.Regexp
	entityFilterReg *regexp.Regexp
)

func init() {
	intervalReg = regexp.MustCompile(intervalPattern)
	entityFilterReg = regexp.MustCompile(entityFilterPattern)
}

func GetBadgeParams(reqPath string, authorizedUser, requestedUser *models.User) (*models.KeyedInterval, *models.Filters, error) {
	isSameUser := authorizedUser != nil && authorizedUser.ID == requestedUser.ID

	var filterEntity, filterKey string
	if groups := entityFilterReg.FindStringSubmatch(reqPath); len(groups) > 2 {
		filterEntity, filterKey = groups[1], groups[2]
	}

	var intervalKey = models.IntervalPast30Days
	if groups := intervalReg.FindStringSubmatch(reqPath); len(groups) > 1 {
		if i, err := helpers.ParseInterval(groups[1]); err == nil {
			intervalKey = i
		}
	}

	_, rangeFrom, rangeTo := helpers.ResolveIntervalTZ(intervalKey, requestedUser.TZ())
	interval := &models.KeyedInterval{
		Interval: models.Interval{Start: rangeFrom, End: rangeTo},
		Key:      intervalKey,
	}

	minStart := rangeTo.AddDate(0, 0, -requestedUser.ShareDataMaxDays)
	// negative value means no limit
	if rangeFrom.Before(minStart) && requestedUser.ShareDataMaxDays >= 0 && !isSameUser {
		return nil, nil, errors.New("requested time range too broad")
	}

	var permitEntity bool
	var filters *models.Filters
	switch filterEntity {
	case "project":
		permitEntity = requestedUser.ShareProjects
		filters = models.NewFiltersWith(models.SummaryProject, filterKey)
	case "os":
		permitEntity = requestedUser.ShareOSs
		filters = models.NewFiltersWith(models.SummaryOS, filterKey)
	case "editor":
		permitEntity = requestedUser.ShareEditors
		filters = models.NewFiltersWith(models.SummaryEditor, filterKey)
	case "language":
		permitEntity = requestedUser.ShareLanguages
		filters = models.NewFiltersWith(models.SummaryLanguage, filterKey)
	case "machine":
		permitEntity = requestedUser.ShareMachines
		filters = models.NewFiltersWith(models.SummaryMachine, filterKey)
	case "label":
		permitEntity = requestedUser.ShareLabels
		filters = models.NewFiltersWith(models.SummaryLabel, filterKey)
		// branches are intentionally omitted here, as only relevant in combination with a project filter
	default:
		// non-entity-specific request, just a general, in-total query
		permitEntity = true
		filters = &models.Filters{}
	}

	if !permitEntity && !isSameUser {
		return nil, nil, errors.New("user did not opt in to share entity-specific data")
	}

	return interval, filters, nil
}
