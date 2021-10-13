package services

import (
	"errors"
	"github.com/leandro-lugaresi/hub"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
	"time"
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

func (srv *ProjectLabelService) GetByUserGrouped(userId string) (map[string][]*models.ProjectLabel, error) {
	labelsByProject := make(map[string][]*models.ProjectLabel)
	userLabels, err := srv.GetByUser(userId)
	if err != nil {
		return nil, err
	}

	for _, l := range userLabels {
		if _, ok := labelsByProject[l.ProjectKey]; !ok {
			labelsByProject[l.ProjectKey] = []*models.ProjectLabel{l}
		} else {
			labelsByProject[l.ProjectKey] = append(labelsByProject[l.ProjectKey], l)
		}
	}
	return labelsByProject, nil
}

func (srv *ProjectLabelService) GetByUserGroupedInverted(userId string) (map[string][]*models.ProjectLabel, error) {
	projectsByLabel := make(map[string][]*models.ProjectLabel)
	userLabels, err := srv.GetByUser(userId)
	if err != nil {
		return nil, err
	}

	for _, l := range userLabels {
		if _, ok := projectsByLabel[l.Label]; !ok {
			projectsByLabel[l.Label] = []*models.ProjectLabel{l}
		} else {
			projectsByLabel[l.Label] = append(projectsByLabel[l.Label], l)
		}
	}
	return projectsByLabel, nil
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
