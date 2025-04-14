package services

import (
	"strings"
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type PluginUserAgentService struct {
	config *config.Config
	cache  *cache.Cache
	db     *gorm.DB
}

func NewPluginUserAgentService(db *gorm.DB) *PluginUserAgentService {
	return &PluginUserAgentService{
		config: config.Get(),
		cache:  cache.New(1*time.Hour, 2*time.Hour),
		db:     db,
	}
}

func formatEditor(editor string) string {
	if strings.ToLower(editor) == "vscode" {
		return "VS Code"
	}
	// TODO: add support for others
	return editor
}

func (srv *PluginUserAgentService) FindOne(attributes models.PluginUserAgent) (*models.PluginUserAgent, error) {
	u := &models.PluginUserAgent{}
	result := srv.db.Where(&attributes).First(u)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, result.Error
	}
	return u, nil
}

func (srv *PluginUserAgentService) CreateOrUpdate(useragent_string, user_id string) (*models.PluginUserAgent, error) {
	useragent, err := models.NewPluginUserAgent(useragent_string, user_id)
	if err != nil {
		return nil, err
	}

	useragent.Editor = formatEditor(useragent.Editor)

	findQuery := models.PluginUserAgent{
		UserID: useragent.UserID,
		Editor: useragent.Editor,
	}
	existing, err := srv.FindOne(findQuery)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		result := srv.db.Model(existing).Updates(map[string]interface{}{"last_seen_at": time.Now()}) // use json tags and avoid this rotten fix.
		if err := result.Error; err != nil {
			return nil, err
		}
	} else {
		result := srv.db.Create(useragent)
		if err := result.Error; err != nil {
			return nil, err
		}
	}

	return useragent, nil
}

func (srv *PluginUserAgentService) FetchUserAgents(userID string) ([]*models.PluginUserAgent, error) {
	var plugins []*models.PluginUserAgent
	if err := srv.db.
		Order("last_seen_at desc").
		Limit(10).
		Where(&models.PluginUserAgent{UserID: userID}).
		Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

type IPluginUserAgentService interface {
	CreateOrUpdate(useragent, user_id string) (*models.PluginUserAgent, error)
	FetchUserAgents(user_id string) ([]*models.PluginUserAgent, error)
}
