package services

import (
	"github.com/duke-git/lancet/v2/condition"
	"github.com/duke-git/lancet/v2/datetime"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"time"
)

const heartbeatPadding = 0 * time.Second

type DurationService struct {
	config           *config.Config
	heartbeatService IHeartbeatService
}

func NewDurationService(heartbeatService IHeartbeatService) *DurationService {
	srv := &DurationService{
		config:           config.Get(),
		heartbeatService: heartbeatService,
	}
	return srv
}

func (srv *DurationService) Get(from, to time.Time, user *models.User, filters *models.Filters) (models.Durations, error) {
	heartbeatsTimeout := user.HeartbeatsTimeout()

	heartbeats, err := srv.heartbeatService.GetAllWithinAsync(from, to, user)
	if err != nil {
		return nil, err
	}

	// Aggregation
	// The below logic is approximately (no filtering, no "same day"-check) equivalent to the SQL query at scripts/aggregate_durations_mysql.sql.
	// A Postgres-compatible script was contributed by @cwilby and is available at scripts/aggregate_durations_postgres.sql
	// I'm hesitant to replicate that logic for sqlite and mssql too (because probably painful to impossible), but we could
	// think about adding a distinction here to use pure-sql aggregation for MySQL and Postgres, and traditional, programmatic
	// aggregation for all other databases.
	var count int
	var latest *models.Duration

	mapping := make(map[string][]*models.Duration)

	for h := range heartbeats {
		d1 := models.NewDurationFromHeartbeat(h).WithEntityIgnored().Hashed()

		// initialize map entry
		if list, ok := mapping[d1.GroupHash]; !ok || len(list) < 1 {
			mapping[d1.GroupHash] = []*models.Duration{}
		}

		// first heartbeat
		if latest == nil {
			mapping[d1.GroupHash] = append(mapping[d1.GroupHash], d1)
			latest = d1
			continue
		}

		// Skip heartbeats that span across two adjacent summaries (assuming there are no more than 1 summary per day).
		// This is relevant to prevent the time difference between generating summaries from raw heartbeats and aggregating pre-generated summaries.
		// For the latter case, the very last heartbeat of a day won't be counted, so we don't want to count it here either
		sameDay := datetime.BeginOfDay(d1.Time.T()) == datetime.BeginOfDay(latest.Time.T())
		dur := condition.Ternary[bool, time.Duration](sameDay, d1.Time.T().Sub(latest.Time.T().Add(latest.Duration)), 0)
		latest.Duration += condition.Ternary[bool, time.Duration](dur < heartbeatsTimeout, dur, heartbeatPadding)

		// Start new "group" if:
		// (a) heartbeats were too far apart each other or,
		// (b) they are of a different entity or,
		// (c) they span across two days
		if dur >= heartbeatsTimeout || latest.GroupHash != d1.GroupHash || !sameDay {
			mapping[d1.GroupHash] = append(mapping[d1.GroupHash], d1)
			latest = d1
		} else {
			latest.NumHeartbeats++
		}

		count++
	}

	durations := make(models.Durations, 0)

	for _, list := range mapping {
		for _, d := range list {
			// Even when filters are applied, we'll still have to compute the whole summary first and then filter out non-matching durations.
			// If we fetched only matching heartbeats in the first place, there will be false positive gaps (see heartbeatsTimeout)
			// in case the user worked on different projects in parallel.
			// See https://github.com/muety/wakapi/issues/535, https://github.com/muety/wakapi/issues/716
			if filters != nil && !filters.MatchDuration(d) {
				continue
			}
			if user.ExcludeUnknownProjects && d.Project == "" {
				continue
			}

			durations = append(durations, d)
		}
	}

	if len(heartbeats) == 1 && len(durations) == 1 {
		durations[0].Duration = heartbeatPadding
	}

	return durations.Sorted(), nil
}
