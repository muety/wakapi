package repositories

import (
	"errors"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"gorm.io/gorm"
)

type ProjectLabelRepository struct {
	config *config.Config
	db     *gorm.DB
}

func NewProjectLabelRepository(db *gorm.DB) *ProjectLabelRepository {
	return &ProjectLabelRepository{config: config.Get(), db: db}
}

func (r *ProjectLabelRepository) GetAll() ([]*models.ProjectLabel, error) {
	var labels []*models.ProjectLabel
	if err := r.db.Find(&labels).Error; err != nil {
		return nil, err
	}
	return labels, nil
}

func (r *ProjectLabelRepository) GetById(id uint) (*models.ProjectLabel, error) {
	label := &models.ProjectLabel{}
	if err := r.db.Where(&models.ProjectLabel{ID: id}).First(label).Error; err != nil {
		return label, err
	}
	return label, nil
}

func (r *ProjectLabelRepository) GetByUser(userId string) ([]*models.ProjectLabel, error) {
	var labels []*models.ProjectLabel
	if err := r.db.
		Where(&models.ProjectLabel{UserID: userId}).
		Find(&labels).Error; err != nil {
		return labels, err
	}
	return labels, nil
}

func (r *ProjectLabelRepository) Insert(label *models.ProjectLabel) (*models.ProjectLabel, error) {
	if !label.IsValid() {
		return nil, errors.New("invalid label")
	}
	result := r.db.Create(label)
	if err := result.Error; err != nil {
		return nil, err
	}
	return label, nil
}

func (r *ProjectLabelRepository) Delete(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(models.ProjectLabel{}).Error
}
