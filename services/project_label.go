package services

import (
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
	"time"
)

type ProjectLabelService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.IProjectLabelRepository
}

func NewProjectLabelService(projectLabelRepository repositories.IProjectLabelRepository) *ProjectLabelService {
	return &ProjectLabelService{
		config:     config.Get(),
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

func (srv *ProjectLabelService) ResolveByUser(userId string) (map[string]string, error) {
	labels := make(map[string]string)
	userLabels, err := srv.GetByUser(userId)
	if err != nil {
		return nil, err
	}

	for _, m := range userLabels {
		labels[m.ProjectKey] = m.Label
	}
	return labels, nil
}

func (srv *ProjectLabelService) Create(label *models.ProjectLabel) (*models.ProjectLabel, error) {
	result, err := srv.repository.Insert(label)
	if err != nil {
		return nil, err
	}

	srv.cache.Delete(result.UserID)
	return result, nil
}

func (srv *ProjectLabelService) Delete(label *models.ProjectLabel) error {
	if label.UserID == "" {
		return errors.New("no user id specified")
	}
	err := srv.repository.Delete(label.ID)
	srv.cache.Delete(label.UserID)
	return err
}
