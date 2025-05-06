package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type DurationService struct {
	config           *config.Config
	heartbeatService IHeartbeatService
	db               *gorm.DB
}

type DurationHeartbeat struct {
	models.Heartbeat // Embed original heartbeat fields
	Duration         time.Duration
	GroupHash        string
	NumHeartbeats    int
}

func NewDurationService(db *gorm.DB) *DurationService {
	heartbeatService := NewHeartbeatService(db)
	srv := &DurationService{
		config:           config.Get(),
		heartbeatService: heartbeatService,
		db:               db,
	}
	return srv
}

func NewTestDurationService(heartbeatService IHeartbeatService) *DurationService {
	srv := &DurationService{
		config:           config.Get(),
		heartbeatService: heartbeatService,
	}
	return srv
}

func (srv *DurationService) Get(from, to time.Time, user *models.User, filters *models.Filters, sliceBy string) (models.Durations, error) {
	if sliceBy == "" {
		sliceBy = SliceByEntity
	}
	heartbeats, err := srv.heartbeatService.GetAllWithin(from, to, user)
	if err != nil {
		return nil, err
	}

	var last_heartbeat_from_yesterday *models.Heartbeat = nil
	var first_heartbeat_from_tomorrow *models.Heartbeat = nil

	if (srv.db != nil) && (user != nil) {

		last_heartbeat_from_yesterday, err = srv.GetYesterdaysLastHeartbeat(user, from)
		if err != nil {
			return nil, err
		}

		first_heartbeat_from_tomorrow, err = srv.FirstHearbeatFromTomorrow(user, from)
		if err != nil {
			return nil, err
		}

	}

	args := models.ProcessHeartbeatsArgs{
		Heartbeats:             heartbeats,
		Start:                  from,
		End:                    to,
		User:                   user,
		LastHeartbeatYesterday: last_heartbeat_from_yesterday,
		FirstHeartbeatTomorrow: first_heartbeat_from_tomorrow,
		SliceBy:                SliceByEntity,
	}

	return srv.MakeDurationsFromHeartbeats(args, filters)
}

func (srv *DurationService) MakeDurationsFromHeartbeats(args models.ProcessHeartbeatsArgs, filters *models.Filters) (models.Durations, error) {
	rawDurations := MakeHeartbeatDurations(args)
	excludeUnknownProjects := args.User.ExcludeUnknownProjects

	durations := make(models.Durations, 0)
	for _, block := range rawDurations {
		d := srv.hashAndFilterDuration(block, excludeUnknownProjects, filters)
		if d != nil {
			durations = append(durations, d)
		}
	}

	if len(args.Heartbeats) == 1 && len(durations) == 1 {
		durations[0].Duration = args.User.HeartbeatsTimeout()
	}

	return durations.Sorted(), nil
}

func (srv *DurationService) FirstHearbeatFromTomorrow(user *models.User, rangeTo time.Time) (*models.Heartbeat, error) {
	startOfTomorrow := rangeTo.Add(time.Second)
	var firstHeartbeatFromTomorrow models.Heartbeat
	result := srv.db.
		Where("user_id = ? AND time >= ?", user.ID, startOfTomorrow).
		Order("time ASC").
		Limit(1).
		Find(&firstHeartbeatFromTomorrow)

	var tomorrowHB *models.Heartbeat = nil
	if result.Error == nil && firstHeartbeatFromTomorrow.ID != 0 {
		tomorrowHB = &firstHeartbeatFromTomorrow
	} else if result.Error != nil && result.Error.Error() != "record not found" {
		config.Log().Error("Failed to retrieve first heartbeat from tomorrow", "error", result.Error)
		return nil, result.Error
	}
	return tomorrowHB, nil
}

func (srv *DurationService) GetYesterdaysLastHeartbeat(user *models.User, rangeFrom time.Time) (*models.Heartbeat, error) {
	var lastHeartbeatFromYesterday models.Heartbeat
	result := srv.db.
		Where("user_id = ? AND time < ?", user.ID, rangeFrom).
		Order("time DESC").
		Limit(1).
		Find(&lastHeartbeatFromYesterday)
	var yesterdayHB *models.Heartbeat = nil
	if result.Error == nil && lastHeartbeatFromYesterday.ID != 0 {
		yesterdayHB = &lastHeartbeatFromYesterday
	} else if result.Error != nil && result.Error.Error() != "record not found" {
		config.Log().Error("Failed to retrieve last heartbeat from yesterday", "error", result.Error)
		return nil, result.Error
	}
	return yesterdayHB, nil
}

func (srv *DurationService) hashAndFilterDuration(dh *models.Duration, excludeUnknownProjects bool, filters *models.Filters) *models.Duration {
	d := dh.WithEntityIgnored().Hashed()

	if excludeUnknownProjects && d.Project == "" {
		return nil
	}

	if filters != nil && !filters.MatchDuration(d) {
		return nil
	}

	if d.Duration == 0 {
		d.Duration = 500 * time.Millisecond
	}
	return d
}
