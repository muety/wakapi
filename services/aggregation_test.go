package services

import (
	"testing"
	"time"

	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
)

func TestAggregationService_GenerateUserJobs_Timezones(t *testing.T) {
	nyTz, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)
	tokyoTz, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	now := time.Now()

	fromUTC := now.AddDate(0, 0, -5).UTC() // pick reference time 5 days ago to ensure multiple jobs are generated

	tests := []struct {
		name     string
		user     *models.User
		tz       *time.Location
		zoneName string
	}{
		{
			name:     "UTC user",
			user:     &models.User{ID: "u_utc", Location: "UTC"},
			tz:       time.UTC,
			zoneName: "UTC",
		},
		{
			name:     "New York user",
			user:     &models.User{ID: "u_ny", Location: "America/New_York"},
			tz:       nyTz,
			zoneName: "America/New_York",
		},
		{
			name:     "Tokyo user",
			user:     &models.User{ID: "u_tokyo", Location: "Asia/Tokyo"},
			tz:       tokyoTz,
			zoneName: "Asia/Tokyo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := generateUserJobs(tt.user, fromUTC)
			assert.NotEmpty(t, jobs)

			endOfYesterday := getStartOfToday(tt.tz).Add(-1 * time.Second)

			for _, job := range jobs {
				assert.Equal(t, tt.user, job.User)

				// boundaries must be in user's timezone
				assert.Equal(t, tt.tz.String(), job.From.Location().String())
				assert.Equal(t, tt.tz.String(), job.To.Location().String())

				// boundaries must be exact midnights (00:00:00) in user's timezone
				assert.Equal(t, 0, job.From.Hour())
				assert.Equal(t, 0, job.From.Minute())
				assert.Equal(t, 0, job.From.Second())
				assert.Equal(t, 0, job.From.Nanosecond())

				assert.Equal(t, 0, job.To.Hour())
				assert.Equal(t, 0, job.To.Minute())
				assert.Equal(t, 0, job.To.Second())
				assert.Equal(t, 0, job.To.Nanosecond())

				// each job must span 1 day (allowing for dst transitions of 23-25 hours)
				diffHours := job.To.Sub(job.From).Hours()
				assert.True(t, diffHours >= 23 && diffHours <= 25)

				// cannot exceed end of yesterday in user's timezone
				assert.True(t, job.From.Before(endOfYesterday))
				assert.True(t, job.To.Before(getStartOfToday(tt.tz)))
			}

			// sequential jobs must be contiguous
			for i := 0; i < len(jobs)-1; i++ {
				assert.True(t, jobs[i].To.Equal(jobs[i+1].From))
			}
		})
	}
}

func TestAggregationService_GetStartOfToday(t *testing.T) {
	tokyoTz, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	nyTz, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)

	todayTokyo := getStartOfToday(tokyoTz)
	todayNY := getStartOfToday(nyTz)

	assert.Equal(t, 0, todayTokyo.Hour())
	assert.Equal(t, 0, todayTokyo.Minute())
	assert.Equal(t, 0, todayTokyo.Second())
	assert.Equal(t, "Asia/Tokyo", todayTokyo.Location().String())

	assert.Equal(t, 0, todayNY.Hour())
	assert.Equal(t, 0, todayNY.Minute())
	assert.Equal(t, 0, todayNY.Second())
	assert.Equal(t, "America/New_York", todayNY.Location().String())
}
