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
		"php":       "PHP 8",
		"blade.php": "Blade",
	}

	sut1, sut2, sut3 := &Heartbeat{
		Entity:   "~/dev/file.py",
		Language: "Python",
	}, &Heartbeat{
		Entity:   "~/dev/file.blade.php",
		Language: "unknown",
	}, &Heartbeat{
		Entity:   "~/dev/file.php",
		Language: "PHP",
	}

	sut1.Augment(testMappings)
	sut2.Augment(testMappings)
	sut3.Augment(testMappings)

	assert.Equal(t, "Python3", sut1.Language)
	assert.Equal(t, "Blade", sut2.Language)
	assert.Equal(t, "PHP 8", sut3.Language)
}

func TestHeartbeat_GetKey(t *testing.T) {
	sut := &Heartbeat{
		Project: "wakapi",
	}

	assert.Equal(t, "wakapi", sut.GetKey(SummaryProject))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryOS))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryMachine))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryCategory))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryLanguage))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryEditor))
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(255))
}

func TestHeartbeat_Hashed(t *testing.T) {
	var sut1, sut2 *Heartbeat

	// same hash if only non-hashed fields are different
	sut1 = &Heartbeat{Entity: "file1", Editor: "vscode", Time: CustomTime(time.Unix(1673810732, 0))}
	sut2 = &Heartbeat{Entity: "file1", Editor: "goland", Time: CustomTime(time.Unix(1673810732, 0))}
	assert.Equal(t, sut1.Hashed().Hash, sut2.Hashed().Hash)

	// different hash if time is different
	sut1 = &Heartbeat{Entity: "file1", Editor: "vscode", Time: CustomTime(time.Unix(1673810732, 0))}
	sut2 = &Heartbeat{Entity: "file1", Editor: "goland", Time: CustomTime(time.Unix(1673810733, 0))}
	assert.NotEqual(t, sut1.Hashed().Hash, sut2.Hashed().Hash)

	// different hash if any other hashed field is different
	sut1 = &Heartbeat{Entity: "file1", Editor: "vscode", Time: CustomTime(time.Unix(1673810732, 0))}
	sut2 = &Heartbeat{Entity: "file2", Editor: "goland", Time: CustomTime(time.Unix(1673810732, 0))}
	assert.NotEqual(t, sut1.Hashed().Hash, sut2.Hashed().Hash)
}

func TestHeartbeat_Hashed_NoCollision(t *testing.T) {
	hashes := map[string]bool{}

	for i := 0; i < 2500; i++ {
		sut := &Heartbeat{
			UserID:   "gopher",
			Entity:   "~/dev/wakapi",
			Type:     "file",
			Category: "coding",
			Project:  "wakapi",
			Branch:   "master",
			Language: "go",
			IsWrite:  false,
			Time:     CustomTime(time.Unix(1673810732+int64(i), 0)),
		}
		assert.NotContains(t, hashes, sut.Hashed().Hash)
		hashes[sut.Hash] = true
	}
}
