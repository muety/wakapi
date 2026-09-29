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

type DurationRepositoryTestSuite struct {
	suite.Suite
	TestUser *models.User
	TestDb   *gorm.DB
	Repo     *DurationRepository
}

func (suite *DurationRepositoryTestSuite) SetupSuite() {
	cfg := config.Empty()
	cfg.Db.Dialect = config.SQLDialectSqlite
	config.Set(cfg)

	db, err := gorm.Open(sqlite.Open(filepath.Join(suite.T().TempDir(), "wakapi_duration_test.db")), &gorm.Config{})
	suite.Require().NoError(err)
	suite.Require().NoError(db.AutoMigrate(&models.User{}, &models.Duration{}))

	suite.TestDb = db
	suite.TestUser = &models.User{ID: "test-user"}
	suite.Require().NoError(db.Create(suite.TestUser).Error)

	suite.Repo = NewDurationRepository(db)
}

func TestDurationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(DurationRepositoryTestSuite))
}

func (suite *DurationRepositoryTestSuite) TestDurationRepository_DeleteByUserAfter_DeletesSpanningDurations() {
	baseTime := time.Date(2026, 9, 29, 10, 0, 0, 0, time.Local)

	durations := []*models.Duration{
		// d1: 09:00 - 09:10 (ended before threshold)
		{
			UserID:   suite.TestUser.ID,
			Time:     models.CustomTime(baseTime.Add(-1 * time.Hour)),
			Duration: 10 * time.Minute,
		},
		// d2: 10:00 - 10:10 (starts before 10:05 threshold, but ends after 10:05)
		{
			UserID:   suite.TestUser.ID,
			Time:     models.CustomTime(baseTime),
			Duration: 10 * time.Minute,
		},
		// d3: 10:15 - 10:20 (starts after 10:05 threshold)
		{
			UserID:   suite.TestUser.ID,
			Time:     models.CustomTime(baseTime.Add(15 * time.Minute)),
			Duration: 5 * time.Minute,
		},
	}

	for _, d := range durations {
		suite.Require().NoError(suite.TestDb.Create(d).Error)
	}

	// Delete after 10:05
	threshold := baseTime.Add(5 * time.Minute)
	err := suite.Repo.DeleteByUserAfter(suite.TestUser, threshold)
	assert.NoError(suite.T(), err)

	remaining, err := suite.Repo.GetAll()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), remaining, 1)
	assert.Equal(suite.T(), baseTime.Add(-1*time.Hour).UnixMilli(), remaining[0].Time.T().UnixMilli())
}
