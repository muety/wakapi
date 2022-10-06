package repositories

import (
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaderboardRepository struct {
	db *gorm.DB
}

func NewLeaderboardRepository(db *gorm.DB) *LeaderboardRepository {
	return &LeaderboardRepository{db: db}
}

func (r *LeaderboardRepository) InsertBatch(items []*models.LeaderboardItem) error {
	if err := r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&items).Error; err != nil {
		return err
	}
	return nil
}

func (r *LeaderboardRepository) CountAllByUser(userId string) (int64, error) {
	var count int64
	err := r.db.
		Table("leaderboard_items").
		Where("user_id = ?", userId).
		Count(&count).Error
	return count, err
}

func (r *LeaderboardRepository) GetAllAggregatedByInterval(key *models.IntervalKey, by *uint8) ([]*models.LeaderboardItem, error) {
	// TODO: distinct by (user, key) to filter out potential duplicates ?
	var items []*models.LeaderboardItem
	q := r.db.
		Select("*, rank() over (partition by \"key\" order by total desc) as \"rank\"").
		Where("\"interval\" in ?", *key)
	q = utils.WhereNullable(q, "\"by\"", by)

	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LeaderboardRepository) GetAggregatedByUserAndInterval(userId string, key *models.IntervalKey, by *uint8) ([]*models.LeaderboardItem, error) {
	var items []*models.LeaderboardItem
	q := r.db.
		Select("*, rank() over (partition by \"key\" order by total desc) as \"rank\"").
		Where("user_id = ?", userId).
		Where("\"interval\" in ?", *key)
	q = utils.WhereNullable(q, "\"by\"", by)

	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LeaderboardRepository) DeleteByUser(userId string) error {
	if err := r.db.
		Where("user_id = ?", userId).
		Delete(models.LeaderboardItem{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *LeaderboardRepository) DeleteByUserAndInterval(userId string, key *models.IntervalKey) error {
	if err := r.db.
		Where("user_id = ?", userId).
		Where("\"interval\" in ?", *key).
		Delete(models.LeaderboardItem{}).Error; err != nil {
		return err
	}
	return nil
}
