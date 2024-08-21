package repositories

import (
	"time"

	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type PluginUserAgentRepository struct {
	db *gorm.DB
}

func NewPluginUserAgentRepository(db *gorm.DB) PluginUserAgentRepository {
	return PluginUserAgentRepository{db: db}
}

func (r *PluginUserAgentRepository) DeleteByIdAndUser(ID, userID string) error {
	if err := r.db.
		Where("id = ?", ID).
		Where("user_id = ?", userID).
		Delete(models.PluginUserAgent{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *PluginUserAgentRepository) FindOne(attributes models.PluginUserAgent) (*models.PluginUserAgent, error) {
	u := &models.PluginUserAgent{}
	result := r.db.Where(&attributes).First(u)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return u, nil
}

func (r *PluginUserAgentRepository) Update(agent *models.PluginUserAgent) (*models.PluginUserAgent, error) {
	result := r.db.Model(agent).Select("editor").Updates(agent)
	if err := result.Error; err != nil {
		return nil, err
	}

	return agent, nil
}

func (r *PluginUserAgentRepository) CreateOrUpdate(useragent_string, user_id string) (*models.PluginUserAgent, error) {
	useragent, err := models.NewPluginUserAgent(useragent_string, user_id)
	if err != nil {
		return nil, err
	}

	findQuery := models.PluginUserAgent{
		Plugin: useragent.Plugin,
		UserID: useragent.UserID,
		Editor: useragent.Editor,
	}
	existing, err := r.FindOne(findQuery)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		result := r.db.Model(existing).Updates(map[string]interface{}{"last_seen_at": time.Now()})
		if err := result.Error; err != nil {
			return nil, err
		}
	} else {
		result := r.db.Create(useragent)
		if err := result.Error; err != nil {
			return nil, err
		}
	}

	return useragent, nil
}

func (r *PluginUserAgentRepository) FetchUserAgents(userID string) ([]*models.PluginUserAgent, error) {
	var plugins []*models.PluginUserAgent
	if err := r.db.
		Order("last_seen_at desc").
		Limit(10).
		Where(&models.PluginUserAgent{UserID: userID}).
		Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

type IPluginUserAgentRepository interface {
	DeleteByIdAndUser(ID, userID string) error
	CreateOrUpdate(useragent_string, user_id string) (*models.PluginUserAgent, error)
	FetchUserAgents(userID string) ([]*models.PluginUserAgent, error)
	Update(agent *models.PluginUserAgent) (*models.PluginUserAgent, error)
}
