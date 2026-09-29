package repositories

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/muety/wakapi/config"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

const (
	TestUserId         = "muety"
	TestCategoryCoding = "coding"
)

type HeartbeatRepositoryTestSuite struct {
	suite.Suite
	TestUser      *models.User
	TestStartTime time.Time
	TestDb        *gorm.DB
}

func (suite *HeartbeatRepositoryTestSuite) SetupSuite() {
	cfg := config.Empty()
	cfg.Db.Dialect = config.SQLDialectSqlite
	config.Set(cfg)

	db, err := gorm.Open(sqlite.Open(filepath.Join(suite.T().TempDir(), "wakapi_test.db")), &gorm.Config{})
	suite.Require().NoError(err)
	suite.Require().NoError(db.AutoMigrate(&models.User{}, &models.Heartbeat{}))

	suite.TestDb = db
	suite.TestUser = &models.User{ID: TestUserId}
	suite.TestStartTime = time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	suite.Require().NoError(db.Create(suite.TestUser).Error)

	// type and category are set via raw column update, because gorm would persist nil values as empty strings instead of NULL
	testHeartbeats := []struct {
		entity   string
		typ      any
		category any
	}{
		{"file_coding", models.HeartbeatTypeFile, TestCategoryCoding},
		{"app_aicoding", models.HeartbeatTypeApp, models.HeartbeatCategoryAiCoding},
		{"app_coding", models.HeartbeatTypeApp, TestCategoryCoding},
		{"file_aicoding", models.HeartbeatTypeFile, models.HeartbeatCategoryAiCoding},
		{"empty_empty", "", ""},
		{"app_null", models.HeartbeatTypeApp, nil},
		{"null_aicoding", nil, models.HeartbeatCategoryAiCoding},
		{"null_null", nil, nil},
		{"file_null", models.HeartbeatTypeFile, nil},
		{"null_coding", nil, TestCategoryCoding},
	}

	for i, h := range testHeartbeats {
		heartbeat := &models.Heartbeat{
			UserID:    TestUserId,
			Entity:    h.entity,
			Time:      models.CustomTime(suite.TestStartTime.Add(time.Duration(i) * time.Minute)),
			CreatedAt: models.CustomTime(suite.TestStartTime),
		}
		heartbeat.Hashed()
		suite.Require().NoError(db.Create(heartbeat).Error)
		suite.Require().NoError(db.Model(heartbeat).UpdateColumns(map[string]any{"type": h.typ, "category": h.category}).Error)
	}
}

func (suite *HeartbeatRepositoryTestSuite) TearDownSuite() {
	if sqlDb, err := suite.TestDb.DB(); err == nil {
		sqlDb.Close()
	}
}

func TestHeartbeatRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatRepositoryTestSuite))
}

func (suite *HeartbeatRepositoryTestSuite) TestHeartbeatRepository_StreamWithinExcludingHeartbeats() {
	entities := suite.streamEntitiesExcluding(models.ExcludeFromDurations)

	assert.NotContains(suite.T(), entities, "app_aicoding")
	assert.Contains(suite.T(), entities, "file_coding")
	assert.Contains(suite.T(), entities, "app_coding")
	assert.Contains(suite.T(), entities, "file_aicoding")
	assert.Contains(suite.T(), entities, "empty_empty")
}

func (suite *HeartbeatRepositoryTestSuite) TestHeartbeatRepository_StreamWithinExcludingHeartbeats_NullValues() {
	entities := suite.streamEntitiesExcluding(models.ExcludeFromDurations)

	assert.Contains(suite.T(), entities, "app_null")
	assert.Contains(suite.T(), entities, "null_aicoding")
	assert.Contains(suite.T(), entities, "null_null")
	assert.Contains(suite.T(), entities, "file_null")
	assert.Contains(suite.T(), entities, "null_coding")
}

func (suite *HeartbeatRepositoryTestSuite) streamEntitiesExcluding(exclusions []models.HeartbeatExclusionFilter) []string {
	sut := NewHeartbeatRepository(suite.TestDb)

	heartbeats, err := sut.StreamWithinExcludingHeartbeats(suite.TestStartTime.Add(-1*time.Hour), suite.TestStartTime.Add(1*time.Hour), suite.TestUser, exclusions)
	suite.Require().NoError(err)

	var entities []string
	for h := range heartbeats {
		entities = append(entities, h.Entity)
	}
	return entities
}
