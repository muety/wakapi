package api

import (
	"github.com/muety/wakapi/mocks"
	"github.com/muety/wakapi/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_fillPlaceholders(t *testing.T) {
	heartbeatServiceMock := new(mocks.HeartbeatServiceMock)
	heartbeatServiceMock.On("GetLatestByUser", mock.Anything).Return(&models.Heartbeat{
		Project: "project1",
	}, nil)

	heartbeatServiceMock.On("GetLatestByFilters", mock.Anything, mock.Anything).Return(&models.Heartbeat{
		Project:  "must not be used",
		Branch:   "replaced2",
		Language: "replaced3",
	}, nil)

	t.Run("when filling placeholders", func(t *testing.T) {
		t.Run("should replace project, language and branch properly", func(t *testing.T) {
			hb := &models.Heartbeat{
				Project:  "<<LAST_PROJECT>>",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
			}
			hb = fillPlaceholders(hb, &models.User{}, heartbeatServiceMock)

			filters1 := heartbeatServiceMock.Calls[1].Arguments.Get(1).(*models.Filters)
			filters2 := heartbeatServiceMock.Calls[2].Arguments.Get(1).(*models.Filters)

			assert.Equal(t, len(heartbeatServiceMock.Calls), 3)
			assert.Equal(t, "project1", filters1.Project[0])
			assert.Equal(t, "project1", filters2.Project[0])
			assert.Equal(t, "project1", hb.Project)
			assert.Equal(t, "replaced2", hb.Branch)
			assert.Equal(t, "replaced3", hb.Language)
		})

		t.Run("should replace nothing if no placeholders given", func(t *testing.T) {
			hb := &models.Heartbeat{
				Project:  "project2",
				Branch:   "branch2",
				Language: "language2",
			}
			hb = fillPlaceholders(hb, &models.User{}, heartbeatServiceMock)
			assert.Equal(t, "project2", hb.Project)
			assert.Equal(t, "branch2", hb.Branch)
			assert.Equal(t, "language2", hb.Language)
		})

		t.Run("should clear placeholders without replacement for browsing heartbeats", func(t *testing.T) {
			hb := &models.Heartbeat{
				Project:  "<<LAST_PROJECT>>",
				Branch:   "<<LAST_BRANCH>>",
				Language: "<<LAST_LANGUAGE>>",
				Type:     "url",
			}
			hb = fillPlaceholders(hb, &models.User{}, heartbeatServiceMock)
			assert.Empty(t, hb.Project)
			assert.Empty(t, hb.Branch)
			assert.Empty(t, hb.Language)
		})
	})
}
