package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/n1try/wakapi/models"
)

type AggregationService struct {
	Db               *sql.DB
	HeartbeatService *HeartbeatService
}

func (srv *AggregationService) Aggregate(from time.Time, to time.Time, user *models.User) {
	heartbeats, err := srv.HeartbeatService.GetAllFrom(from, user)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range heartbeats {
		fmt.Printf("%+v\n", h)
	}
}

func (srv *AggregationService) aggregateBy(*[]models.Heartbeat, models.AggregationType) *models.Aggregation {
	return &models.Aggregation{}
}
