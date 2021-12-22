package services

import (
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"time"
)

const HeartbeatDiffThreshold = 2 * time.Minute

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

func (srv *DurationService) Get(from, to time.Time, user *models.User) (models.Durations, error) {
	heartbeats, err := srv.heartbeatService.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}

	// Aggregation
	var count int
	var latest *models.Duration

	mapping := make(map[string][]*models.Duration)

	for _, h := range heartbeats {
		d1 := models.NewDurationFromHeartbeat(h)

		if list, ok := mapping[d1.GroupHash]; !ok || len(list) < 1 {
			mapping[d1.GroupHash] = []*models.Duration{d1}
		}

		if latest == nil {
			latest = d1
			continue
		}

		dur := d1.Time.T().Sub(latest.Time.T().Add(latest.Duration))
		if dur > HeartbeatDiffThreshold {
			dur = HeartbeatDiffThreshold
		}
		latest.Duration += dur

		if dur >= HeartbeatDiffThreshold || latest.GroupHash != d1.GroupHash {
			list := mapping[d1.GroupHash]
			if d0 := list[len(list)-1]; d0 != d1 {
				mapping[d1.GroupHash] = append(mapping[d1.GroupHash], d1)
			}
			latest = d1
		} else {
			latest.NumHeartbeats++
		}

		count++
	}

	durations := make(models.Durations, 0, count)

	for _, list := range mapping {
		for _, d := range list {
			if d.Duration == 0 {
				d.Duration = HeartbeatDiffThreshold
			}
			durations = append(durations, d)
		}
	}

	return durations.Sorted(), nil
}
