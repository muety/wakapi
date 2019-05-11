package services

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
)

type AggregationService struct {
	Db               *gorm.DB
	HeartbeatService *HeartbeatService
}

func (srv *AggregationService) Aggregate(from time.Time, to time.Time, user *models.User) {
	heartbeats, err := srv.HeartbeatService.GetAllFrom(from, user)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range *heartbeats {
		fmt.Printf("%+v\n", h)
	}
}

func (srv *AggregationService) aggregateBy(*[]models.Heartbeat, models.AggregationType) *models.Aggregation {
	return &models.Aggregation{}
}
