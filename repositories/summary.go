package repositories

import (
	"errors"
	"log/slog"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SummaryRepository struct {
	BaseRepository
}

func NewSummaryRepository(db *gorm.DB) *SummaryRepository {
	return &SummaryRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *SummaryRepository) GetAll() ([]*models.Summary, error) {
	var summaries []*models.Summary
	if err := r.db.
		Order("from_time asc").
		// branch summaries are currently not persisted, as only relevant in combination with project filter
		Find(&summaries).Error; err != nil {
		return nil, err
	}

	if err := r.populateItems(summaries, []clause.Interface{}); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *SummaryRepository) InsertWithRetry(summary *models.Summary) (err error) {
	// in case of duplicate key error, retry inserting up to three times
	// https://github.com/muety/wakapi/issues/877
	for i := 0; i < 3; i++ {
		err = r.Insert(summary)
		if err == nil || !errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
		slog.Warn("retrying to insert summary", "userId", summary.UserID, "fromTime", summary.FromTime.T(), "toTime", summary.ToTime.T(), "error", err)
		time.Sleep(1 * time.Second)
	}
	return err
}

func (r *SummaryRepository) Insert(summary *models.Summary) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(summary).Error; err != nil {
			return err
		}

		itemsToCreate := []*models.SummaryItem{}

		// required due to setting gorm:"-" in the model definition
		// see https://github.com/muety/wakapi/issues/600#issuecomment-1921723789
		// see https://github.com/muety/wakapi/pull/592#discussion_r1450478355
		for _, item := range summary.Machines {
			item.SummaryID = summary.ID
			itemsToCreate = append(itemsToCreate, item)
		}

		for _, item := range summary.Languages {
			item.SummaryID = summary.ID
			itemsToCreate = append(itemsToCreate, item)
		}

		for _, item := range summary.OperatingSystems {
			item.SummaryID = summary.ID
			itemsToCreate = append(itemsToCreate, item)
		}

		for _, item := range summary.Editors {
			item.SummaryID = summary.ID
			itemsToCreate = append(itemsToCreate, item)
		}

		for _, item := range summary.Categories {
			item.SummaryID = summary.ID
			itemsToCreate = append(itemsToCreate, item)
		}

		if len(itemsToCreate) > 0 {
			if err := tx.Create(itemsToCreate).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *SummaryRepository) GetByUserWithin(user *models.User, from, to time.Time) ([]*models.Summary, error) {
	var summaries []*models.Summary

	queryConditions := []clause.Interface{
		clause.Where{Exprs: r.db.Statement.BuildCondition("user_id = ?", user.ID)},
		clause.Where{Exprs: r.db.Statement.BuildCondition("from_time >= ?", from.Local())},
		clause.Where{Exprs: r.db.Statement.BuildCondition("to_time <= ?", to.Local())},
	}

	q := r.db.Model(&models.Summary{}).
		Order("from_time asc")

	for _, c := range queryConditions {
		q.Statement.AddClause(c)
	}

	// branch summaries are currently not persisted, as only relevant in combination with project filter
	if err := q.Find(&summaries).Error; err != nil {
		return nil, err
	}

	if err := r.populateItems(summaries, queryConditions); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *SummaryRepository) GetLastByUser() ([]*models.TimeByUser, error) {
	var result []*models.TimeByUser
	r.db.Model(&models.User{}).
		Select(utils.QuoteSql(r.db, "users.id as %s, max(to_time) as time", "user")).
		Joins("left join summaries on users.id = summaries.user_id").
		Group("users.id").
		Scan(&result)
	return result, nil
}

func (r *SummaryRepository) DeleteByUser(userId string) error {
	if err := r.db.
		Where("user_id = ?", userId).
		Delete(models.Summary{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *SummaryRepository) DeleteByUserBefore(userId string, t time.Time) error {
	if err := r.db.
		Where("user_id = ?", userId).
		Where("to_time <= ?", t.Local()).
		Delete(models.Summary{}).Error; err != nil {
		return err
	}
	return nil
}

// inplace
func (r *SummaryRepository) populateItems(summaries []*models.Summary, conditions []clause.Interface) error {
	var items []*models.SummaryItem

	summaryMap := slice.GroupWith[*models.Summary, uint](summaries, func(s *models.Summary) uint {
		return s.ID
	})

	q := r.db.Model(&models.SummaryItem{}).
		Select("summary_items.*").
		Joins("cross join summaries").
		Where("summary_items.summary_id = summaries.id").
		Where("num_heartbeats > ?", 0)

	for _, c := range conditions {
		q.Statement.AddClause(c)
	}

	if err := q.Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		if _, ok := summaryMap[item.SummaryID]; !ok {
			continue
		}
		l := summaryMap[item.SummaryID][0].GetByType(item.Type)
		*l = append(*l, item)
	}

	return nil
}
