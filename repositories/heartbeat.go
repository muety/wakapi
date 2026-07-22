package repositories

import (
	"database/sql"
	"time"

	"github.com/duke-git/lancet/v2/condition"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

type HeartbeatRepository struct {
	BaseRepository
	config *conf.Config
}

func NewHeartbeatRepository(db *gorm.DB) *HeartbeatRepository {
	return &HeartbeatRepository{BaseRepository: NewBaseRepository(db), config: conf.Get()}
}

// Use with caution!!
func (r *HeartbeatRepository) GetAll() ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := r.db.Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) InsertBatch(heartbeats []*models.Heartbeat) error {
	return InsertBatchChunked[*models.Heartbeat](heartbeats, &models.Heartbeat{}, r.db)
}

func (r *HeartbeatRepository) GetLatestByUser(user *models.User) (*models.Heartbeat, error) {
	var heartbeat models.Heartbeat
	q := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{UserID: user.ID}).
		Limit(1)
	q = r.queryAddTimeSorting(q, true)
	if err := q.Scan(&heartbeat).Error; err != nil {
		return nil, err
	}
	return &heartbeat, nil
}

func (r *HeartbeatRepository) GetLatestByOriginAndUser(origin string, user *models.User) (*models.Heartbeat, error) {
	var heartbeat models.Heartbeat
	q := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{
			UserID: user.ID,
			Origin: origin,
		}).
		Limit(1)
	q = r.queryAddTimeSorting(q, true)
	if err := q.Scan(&heartbeat).Error; err != nil {
		return nil, err
	}
	return &heartbeat, nil
}

func (r *HeartbeatRepository) GetWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	// https://stackoverflow.com/a/20765152/3112139
	var heartbeats []*models.Heartbeat
	if err := r.buildTimeFilteredQuery(user.ID, from.Local(), to.Local()).Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) StreamWithin(from, to time.Time, user *models.User) (chan *models.Heartbeat, error) {
	out := make(chan *models.Heartbeat)

	rows, err := r.buildTimeFilteredQuery(user.ID, from.Local(), to.Local()).Rows()
	if err != nil {
		return nil, err
	}

	go streamRows[models.Heartbeat](rows, out, r.db, func(err error) {
		conf.Log().Error("failed to scan heartbeats row", "user", user.ID, "from", from, "to", to, "error", err)
	})
	return out, nil
}

func (r *HeartbeatRepository) StreamWithinBatched(from, to time.Time, user *models.User, batchSize int) (chan []*models.Heartbeat, error) {
	out := make(chan []*models.Heartbeat)

	rows, err := r.buildTimeFilteredQuery(user.ID, from.Local(), to.Local()).Rows()
	if err != nil {
		return nil, err
	}

	go streamRowsBatched[models.Heartbeat](rows, out, r.db, batchSize, func(err error) {
		conf.Log().Error("failed to scan heartbeats row", "user", user.ID, "from", from, "to", to, "error", err)
	})
	return out, nil
}

func (r *HeartbeatRepository) GetAllWithinByFilters(from, to time.Time, user *models.User, filterMap map[string][]string) ([]*models.Heartbeat, error) {
	// https://stackoverflow.com/a/20765152/3112139
	var heartbeats []*models.Heartbeat

	q := r.buildTimeFilteredQuery(user.ID, from.Local(), to.Local())
	q = filteredQuery(q, filterMap)

	if err := q.Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) StreamWithinByFilters(from, to time.Time, user *models.User, filterMap map[string][]string) (chan *models.Heartbeat, error) {
	out := make(chan *models.Heartbeat)

	q := r.buildTimeFilteredQuery(user.ID, from.Local(), to.Local())
	q = filteredQuery(q, filterMap)

	rows, err := q.Rows()
	if err != nil {
		return nil, err
	}

	go streamRows[models.Heartbeat](rows, out, r.db, func(err error) {
		conf.Log().Error("failed to scan filtered heartbeats row", "user", user.ID, "from", from, "to", to, "error", err)
	})

	return out, nil
}

func (r *HeartbeatRepository) GetLatestByFilters(user *models.User, filterMap map[string][]string) (*models.Heartbeat, error) {
	var heartbeat *models.Heartbeat

	q := r.db.Model(&models.Heartbeat{}).Where(&models.Heartbeat{UserID: user.ID})
	q = r.queryAddTimeSorting(q, true)
	q = filteredQuery(q, filterMap)

	if err := q.Limit(1).Scan(&heartbeat).Error; err != nil {
		return nil, err
	}
	return heartbeat, nil
}

func (r *HeartbeatRepository) GetFirstAll() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	err := r.db.Raw("select user_id as user, first as time from user_heartbeats_range").Scan(&result).Error
	return result, err
}

func (r *HeartbeatRepository) GetLastAll() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	err := r.db.Raw("select user_id as user, last as time from user_heartbeats_range").Scan(&result).Error
	return result, err
}

func (r *HeartbeatRepository) GetRangeByUser(user *models.User) (*models.RangeByUser, error) {
	var result *models.RangeByUser
	err := r.db.Raw("select user_id as user, first, last from user_heartbeats_range where user_id = ?", user.ID).Scan(&result).Error
	return result, err
}

func (r *HeartbeatRepository) Count(approximate bool) (count int64, err error) {
	if r.config.Db.IsMySQL() && approximate {
		err = r.db.Table("information_schema.tables").
			Select("table_rows").
			Where("table_schema = ?", r.config.Db.Name).
			Where("table_name = 'heartbeats'").
			Scan(&count).Error
	}

	if count == 0 {
		err = r.db.
			Model(&models.Heartbeat{}).
			Count(&count).Error
	}
	return count, nil
}

func (r *HeartbeatRepository) CountByUser(user *models.User) (int64, error) {
	var count int64
	if err := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{UserID: user.ID}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *HeartbeatRepository) CountByUsers(users []*models.User) ([]*models.CountByUser, error) {
	var counts []*models.CountByUser

	userIds := make([]string, len(users))
	for i, u := range users {
		userIds[i] = u.ID
	}

	if len(userIds) == 0 {
		return counts, nil
	}

	if err := r.db.
		Model(&models.Heartbeat{}).
		Select(utils.QuoteSql(r.db, "user_id as %s, count(id) as %s", "user", "count")).
		Where("user_id in ?", userIds).
		Group("user").
		Find(&counts).Error; err != nil {
		return counts, err
	}

	return counts, nil
}

func (r *HeartbeatRepository) GetEntitySetByUser(entityType uint8, userId string) ([]string, error) {
	var results []string
	if err := r.db.
		Model(&models.Heartbeat{}).
		Distinct(models.GetEntityColumn(entityType)).
		Where(&models.Heartbeat{UserID: userId}).
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *HeartbeatRepository) DeleteBefore(t time.Time) error {
	q := r.queryAddTimeFilterLessEqual(r.db.Model(models.Heartbeat{}), t.Local())
	if err := q.Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) DeleteByUser(user *models.User) error {
	if err := r.db.
		Where("user_id = ?", user.ID).
		Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) DeleteByUserBefore(user *models.User, t time.Time) error {
	q := r.queryAddTimeFilterLessEqual(r.db.Model(models.Heartbeat{}), t.Local())
	if err := q.
		Where("user_id = ?", user.ID).
		Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) GetUserProjectStats(user *models.User, from, to time.Time) ([]*models.ProjectStats, error) {
	var projectStats []*models.ProjectStats

	// note: query takes quite long, depending on the number of heartbeats (~ 7 seconds for ~ 500k heartbeats)
	// TODO: refactor this to use summaries once we implemented persisting filtered, multi-interval summaries
	// see https://github.com/muety/wakapi/issues/524#issuecomment-1731668391

	// multi-line string with backticks yields an error with the github.com/glebarez/sqlite driver

	args := []interface{}{
		sql.Named("userid", user.ID),
		sql.Named("from", models.CustomTime(from)),
		sql.Named("to", models.CustomTime(to)),
	}

	query := "with lang_grouped as (" +
		"select project, language, count(*) as lang_cnt, min(time) as lang_min, max(time) as lang_max " +
		"from heartbeats " +
		"where user_id = @userid" +
		" and project != ''" +
		" and time between @from and @to" +
		" and language is not null and language != '' " +
		"group by project, language" +
		"), ranked as (" +
		"select project, language, lang_cnt, lang_min, lang_max, " +
		"sum(lang_cnt) over (partition by project) as total_cnt, " +
		"min(lang_min) over (partition by project) as overall_first, " +
		"max(lang_max) over (partition by project) as overall_last, " +
		"row_number() over (partition by project order by lang_cnt desc) as rn " +
		"from lang_grouped" +
		") select project, overall_first as first, overall_last as last, " +
		"total_cnt as count, language as top_language, " +
		"@userid as user_id " +
		"from ranked " +
		"where rn = 1"

	if err := r.db.
		Raw(query, args...).
		Scan(&projectStats).Error; err != nil {
		return nil, err
	}

	return projectStats, nil
}

func (r *HeartbeatRepository) GetUserAgentsByUser(user *models.User) ([]*models.UserAgent, error) {
	var results []*models.UserAgent
	if err := r.db.
		Model(&models.Heartbeat{}).
		Select("user_agent as value, operating_system as os, editor, min(time) as first_seen, max(time) as last_seen").
		Where(&models.Heartbeat{UserID: user.ID}).
		Not("user_agent = ''").
		Group("user_agent, operating_system, editor").
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *HeartbeatRepository) buildTimeFilteredQuery(userId string, from, to time.Time) *gorm.DB {
	query := r.db.Model(&models.Heartbeat{}).Where(&models.Heartbeat{UserID: userId})
	query = r.queryAddTimeFilterBetween(query, from, to)
	query = r.queryAddTimeSorting(query, false)
	return query
}

func (r *HeartbeatRepository) queryAddTimeFilterBetween(q *gorm.DB, from, to time.Time) *gorm.DB {
	return q.
		Where("time >= ?", models.CustomTime(from.Local())).
		Where("time < ?", models.CustomTime(to.Local()))
}

func (r *HeartbeatRepository) queryAddTimeFilterLessEqual(q *gorm.DB, t time.Time) *gorm.DB {
	return q.Where("time <= ?", models.CustomTime(t.Local()))
}

func (r *HeartbeatRepository) queryAddTimeSorting(q *gorm.DB, desc bool) *gorm.DB {
	order := condition.Ternary(desc, "desc", "asc")
	return q.Order("time " + order)
}
