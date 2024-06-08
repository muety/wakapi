package services

import (
	"time"

	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/repositories"
	"github.com/patrickmn/go-cache"
)

type GoalService struct {
	config     *config.Config
	cache      *cache.Cache
	repository repositories.IGoalRepository
}

func NewGoalService(goalRepo repositories.IGoalRepository) *GoalService {
	return &GoalService{
		config:     config.Get(),
		cache:      cache.New(1*time.Hour, 2*time.Hour),
		repository: goalRepo,
	}
}

func (srv *GoalService) Create(newGaol *models.Goal) (*models.Goal, error) {
	return srv.repository.Create(newGaol)
}

func (srv *GoalService) GetGoalForUser(id, userID string) (*models.Goal, error) {
	return srv.repository.GetByIdForUser(id, userID)
}

func (srv *GoalService) DeleteGoal(id, userID string) error {
	return srv.repository.DeleteByIdAndUser(id, userID)
}

func (srv *GoalService) FetchUserGoals(id string) ([]*models.Goal, error) {
	return srv.repository.FetchUserGoals(id)
}
