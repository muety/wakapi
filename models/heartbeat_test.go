package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHeartbeat_Valid_Success(t *testing.T) {
	sut := &Heartbeat{
		User: &User{
			ID: "johndoe@example.org",
		},
		UserID: "johndoe@example.org",
		Time:   CustomTime(time.Now()),
	}
	assert.True(t, sut.Valid())
}

func TestHeartbeat_Valid_MissingUser(t *testing.T) {
	sut := &Heartbeat{
		Time: CustomTime(time.Now()),
	}
	assert.False(t, sut.Valid())
}

func TestHeartbeat_Augment(t *testing.T) {
	testMappings := map[string]string{
		"py":        "Python3",
		"foo":       "Foo Script",
		"blade.php": "Blade",
	}

	sut1, sut2 := &Heartbeat{
		Entity:   "~/dev/file.py",
		Language: "Python",
	}, &Heartbeat{
		Entity:   "~/dev/file.blade.php",
		Language: "unknown",
	}

	sut1.Augment(testMappings)
	sut2.Augment(testMappings)

	assert.Equal(t, "Python3", sut1.Language)
	assert.Equal(t, "Blade", sut2.Language)
}

func TestHeartbeat_GetKey(t *testing.T) {
	sut := &Heartbeat{
		Project: "wakapi",
	}

	assert.Equal(t, "wakapi", sut.GetKey(SummaryProject))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryOS))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryMachine))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryLanguage))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryEditor))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(255))
}
