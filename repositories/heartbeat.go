package repositories

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HeartbeatRepository struct {
	db     *gorm.DB
	config *conf.Config
}

func NewHeartbeatRepository(db *gorm.DB) *HeartbeatRepository {
	return &HeartbeatRepository{config: conf.Get(), db: db}
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
	if err := r.db.
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&heartbeats).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) GetLatestByUser(user *models.User) (*models.Heartbeat, error) {
	var heartbeat models.Heartbeat
	if err := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{UserID: user.ID}).
		Order("time desc").
		First(&heartbeat).Error; err != nil {
		return nil, err
	}
	return &heartbeat, nil
}

func (r *HeartbeatRepository) GetLatestByOriginAndUser(origin string, user *models.User) (*models.Heartbeat, error) {
	var heartbeat models.Heartbeat
	if err := r.db.
		Model(&models.Heartbeat{}).
		Where(&models.Heartbeat{
			UserID: user.ID,
			Origin: origin,
		}).
		Order("time desc").
		First(&heartbeat).Error; err != nil {
		return nil, err
	}
	return &heartbeat, nil
}

func (r *HeartbeatRepository) GetAllWithin(from, to time.Time, user *models.User) ([]*models.Heartbeat, error) {
	// https://stackoverflow.com/a/20765152/3112139
	var heartbeats []*models.Heartbeat
	if err := r.db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Where("time >= ?", from.Local()).
		Where("time < ?", to.Local()).
		Order("time asc").
		Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) GetAllWithinByFilters(from, to time.Time, user *models.User, filterMap map[string][]string) ([]*models.Heartbeat, error) {
	// https://stackoverflow.com/a/20765152/3112139
	var heartbeats []*models.Heartbeat

	q := r.db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Where("time >= ?", from.Local()).
		Where("time < ?", to.Local()).
		Order("time asc")
	q = r.filteredQuery(q, filterMap)

	if err := q.Find(&heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

func (r *HeartbeatRepository) GetLatestByFilters(user *models.User, filterMap map[string][]string) (*models.Heartbeat, error) {
	var heartbeat *models.Heartbeat

	q := r.db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Order("time desc")
	q = r.filteredQuery(q, filterMap)

	if err := q.First(&heartbeat).Error; err != nil {
		return nil, err
	}
	return heartbeat, nil
}

func (r *HeartbeatRepository) GetFirstByUsers() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	r.db.Model(&models.User{}).
		Select(fmt.Sprintf("users.id as %s, min(time) as %s", utils.QuoteDbIdentifier(r.db, "user"), utils.QuoteDbIdentifier(r.db, "time"))).
		Joins("left join heartbeats on users.id = heartbeats.user_id").
		Group("users.id").
		Scan(&result)
	return result, nil
}

func (r *HeartbeatRepository) GetLastByUsers() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	r.db.Model(&models.User{}).
		Select(fmt.Sprintf("users.id as %s, max(time) as %s", utils.QuoteDbIdentifier(r.db, "user"), utils.QuoteDbIdentifier(r.db, "time"))).
		Joins("left join heartbeats on users.id = heartbeats.user_id").
		Group("user").
		Scan(&result)
	return result, nil
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
		Select(fmt.Sprintf("user_id as %s, count(id) as %s", utils.QuoteDbIdentifier(r.db, "user"), utils.QuoteDbIdentifier(r.db, "count"))).
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
	if err := r.db.
		Where("time <= ?", t.Local()).
		Delete(models.Heartbeat{}).Error; err != nil {
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
	if err := r.db.
		Where("user_id = ?", user.ID).
		Where("time <= ?", t.Local()).
		Delete(models.Heartbeat{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *HeartbeatRepository) GetUserProjectStats(user *models.User, from, to time.Time, limit, offset int) ([]*models.ProjectStats, error) {
	var projectStats []*models.ProjectStats

	// note: limit / offset doesn't really improve query performance
	// query takes quite long, depending on the number of heartbeats (~ 7 seconds for ~ 500k heartbeats)
	// TODO: refactor this to use summaries once we implemented persisting filtered, multi-interval summaries
	// see https://github.com/muety/wakapi/issues/524#issuecomment-1731668391

	// multi-line string with backticks yields an error with the github.com/glebarez/sqlite driver
	if err := r.db.
		Raw("with projects as ( "+
			"select project, user_id, min(time) as first, max(time) as last, count(*) as cnt "+
			"from heartbeats "+
			"where user_id = ? and project != '' "+
			"and time between ? and ? "+
			"and language is not null and language != '' and project != '' "+
			"group by project, user_id "+
			"order by last desc "+
			"limit ? offset ? "+
			") "+
			"select distinct project, min(first) as first, min(last) as last, min(cnt) as count, first_value(language) over (partition by project order by count(*) desc) as top_language "+
			"from heartbeats "+
			"inner join projects using (project, user_id) "+
			"group by project, language "+
			"order by last desc", user.ID, from, to, limit, offset).
		Scan(&projectStats).Error; err != nil {
		return nil, err
	}

	return projectStats, nil
}

func (r *HeartbeatRepository) filteredQuery(q *gorm.DB, filterMap map[string][]string) *gorm.DB {
	for col, vals := range filterMap {
		q = q.Where(col+" in ?", slice.Map[string, string](vals, func(i int, val string) string {
			// query for "unknown" projects, languages, etc.
			if val == "-" {
				return ""
			}
			return val
		}))
	}
	return q
}
