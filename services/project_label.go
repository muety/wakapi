package services

import (
	"errors"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/leandro-lugaresi/hub"
	"github.com/patrickmn/go-cache"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
)

type ProjectLabelService struct {
	config     *config.Config
	cache      *cache.Cache
	eventBus   *hub.Hub
	repository repositories.IProjectLabelRepository
}

func NewProjectLabelService(projectLabelRepository repositories.IProjectLabelRepository) *ProjectLabelService {
	return &ProjectLabelService{
		config:     config.Get(),
		eventBus:   config.EventBus(),
		repository: projectLabelRepository,
		cache:      cache.New(24*time.Hour, 24*time.Hour),
	}
}

func (srv *ProjectLabelService) GetById(id uint) (*models.ProjectLabel, error) {
	return srv.repository.GetById(id)
}

func (srv *ProjectLabelService) GetByUser(userId string) ([]*models.ProjectLabel, error) {
	if labels, found := srv.cache.Get(userId); found {
		return labels.([]*models.ProjectLabel), nil
	}

	labels, err := srv.repository.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	srv.cache.Set(userId, labels, cache.DefaultExpiration)
	return labels, nil
}

// GetByUserGrouped returns lists of project labels, grouped by their project key
func (srv *ProjectLabelService) GetByUserGrouped(userId string) (map[string][]*models.ProjectLabel, error) {
	userLabels, err := srv.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	mappedLabels := slice.GroupWith[*models.ProjectLabel, string](userLabels, func(l *models.ProjectLabel) string {
		return l.ProjectKey
	})
	return mappedLabels, nil
}

// GetByUserGroupedInverted returns lists of project labels, grouped by their label key
func (srv *ProjectLabelService) GetByUserGroupedInverted(userId string) (map[string][]*models.ProjectLabel, error) {
	userLabels, err := srv.GetByUser(userId)
	if err != nil {
		return nil, err
	}
	mappedLabels := slice.GroupWith[*models.ProjectLabel, string](userLabels, func(l *models.ProjectLabel) string {
		return l.Label
	})
	return mappedLabels, nil
}

func (srv *ProjectLabelService) Create(label *models.ProjectLabel) (*models.ProjectLabel, error) {
	result, err := srv.repository.Insert(label)
	if err != nil {
		return nil, err
	}

	srv.cache.Delete(result.UserID)
	srv.notifyUpdate(label, false)
	return result, nil
}

func (srv *ProjectLabelService) Delete(label *models.ProjectLabel) error {
	if label.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(label.ID)
	srv.cache.Delete(label.UserID)
	srv.notifyUpdate(label, true)
	return err
}

func (srv *ProjectLabelService) notifyUpdate(label *models.ProjectLabel, isDelete bool) {
	name := config.EventProjectLabelCreate
	if isDelete {
		name = config.EventProjectLabelDelete
	}
	srv.eventBus.Publish(hub.Message{
		Name:   name,
		Fields: map[string]interface{}{config.FieldPayload: label, config.FieldUserId: label.UserID},
	})
}
