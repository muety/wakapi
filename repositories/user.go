package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/condition"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
)

type UserRepository struct {
	BaseRepository
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *UserRepository) FindOne(attributes models.User) (*models.User, error) {
	u := &models.User{}
	if err := r.db.Where(&attributes).First(u).Error; err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) GetByIds(userIds []string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Model(&models.User{}).
		Where("id in ?", userIds).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Where(&models.User{}).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetMany(ids []string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.
		Table("users").
		Where("id in ?", ids).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAllByReports(reportsEnabled bool) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.Where(&models.User{ReportsWeekly: reportsEnabled}).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAllByLeaderboard(leaderboardEnabled bool) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.Where(&models.User{PublicLeaderboard: leaderboardEnabled}).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetByLoggedInAfter(t time.Time) ([]*models.User, error) {
	return r.getByLoggedIn(t, true)
}

func (r *UserRepository) GetByLoggedInBefore(t time.Time) ([]*models.User, error) {
	return r.getByLoggedIn(t, false)
}

// Returns a list of user ids, whose last heartbeat is not older than t
// NOTE: Only ID field will be populated
func (r *UserRepository) GetByLastActiveAfter(t time.Time) ([]*models.User, error) {
	subQuery1 := r.db.Model(&models.Heartbeat{}).
		Select(utils.QuoteSql(r.db, "user_id as %s, max(time) as time", "user")).
		Group("user_id")

	var userIds []string
	if err := r.db.
		Select(utils.QuoteSql(r.db, "user as %s", "id")).
		Table("(?) as q", subQuery1).
		Where("time >= ?", t.Local()).
		Scan(&userIds).Error; err != nil {
		return nil, err
	}

	return r.GetByIds(userIds)
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.
		Model(&models.User{}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) InsertOrGet(user *models.User) (*models.User, bool, error) {
	// no need to replace "" with nil for nullable string columns, because they'll have "default:null" set
	// this hacky replacement is only needed for updates

	if u, err := r.FindOne(models.User{ID: user.ID}); err == nil && u != nil && u.ID != "" {
		return u, false, nil
	}

	result := r.db.Create(user)
	if err := result.Error; err != nil {
		return nil, false, err
	}

	return user, true, nil
}

func (r *UserRepository) Update(user *models.User) (*models.User, error) {
	// for string columns with a unique index we must ensure null is used instead of empty string
	// proper way would be to use *string or sql.NullString member type, we're doing it the hacky way though
	// see https://stackoverflow.com/a/54699204/3112139

	updateMap := map[string]interface{}{
		"api_key":                  user.ApiKey,
		"password":                 user.Password,
		"email":                    condition.Ternary[bool, interface{}](user.Email == "", nil, user.Email),
		"last_logged_in_at":        user.LastLoggedInAt,
		"share_data_max_days":      user.ShareDataMaxDays,
		"share_editors":            user.ShareEditors,
		"share_languages":          user.ShareLanguages,
		"share_oss":                user.ShareOSs,
		"share_projects":           user.ShareProjects,
		"share_machines":           user.ShareMachines,
		"share_labels":             user.ShareLabels,
		"share_activity_chart":     user.ShareActivityChart,
		"wakatime_api_key":         user.WakatimeApiKey,
		"wakatime_api_url":         user.WakatimeApiUrl,
		"has_data":                 user.HasData,
		"reset_token":              user.ResetToken,
		"unsubscribe_token":        user.UnsubscribeToken,
		"location":                 user.Location,
		"start_of_week":            user.StartOfWeek,
		"reports_weekly":           user.ReportsWeekly,
		"public_leaderboard":       user.PublicLeaderboard,
		"subscribed_until":         user.SubscribedUntil,
		"subscription_renewal":     user.SubscriptionRenewal,
		"stripe_customer_id":       user.StripeCustomerId,
		"invited_by":               user.InvitedBy,
		"exclude_unknown_projects": user.ExcludeUnknownProjects,
		"heartbeats_timeout_sec":   user.HeartbeatsTimeoutSec,
		"readme_stats_base_url":    user.ReadmeStatsBaseUrl,
		"auth_type":                user.AuthType,
		"sub":                      condition.Ternary[bool, interface{}](user.Sub == "", nil, user.Sub),
	}

	result := r.db.Model(user).Updates(updateMap)
	if err := result.Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateField(user *models.User, key string, value interface{}) (*models.User, error) {
	// for string columns with a unique index we must ensure null is used instead of empty string
	// proper way would be to use *string or sql.NullString member type, we're doing it the hacky way though
	// see https://stackoverflow.com/a/54699204/3112139

	if strVal, ok := value.(string); ok && key == "email" && strVal == "" {
		value = nil
	}
	if strVal, ok := value.(string); ok && key == "sub" && strVal == "" {
		value = nil
	}

	result := r.db.Model(user).Update(key, value)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected != 1 {
		return nil, errors.New("nothing updated")
	}

	return user, nil
}

func (r *UserRepository) Delete(user *models.User) error {
	return r.DeleteTx(user, r.db)
}

func (r *UserRepository) DeleteTx(user *models.User, tx *gorm.DB) error {
	return tx.Delete(user).Error
}

func (r *UserRepository) getByLoggedIn(t time.Time, after bool) ([]*models.User, error) {
	var users []*models.User
	comparator := condition.Ternary[bool, string](after, ">=", "<=")
	if err := r.db.
		Where(fmt.Sprintf("last_logged_in_at %s ?", comparator), t.Local()).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
