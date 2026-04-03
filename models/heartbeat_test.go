package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	now := time.Unix(1600000000, 0)
	h1 := &Heartbeat{
		UserID:          "user1",
		Entity:          "entity1",
		Type:            "file",
		Category:        "coding",
		Project:         "project1",
		Branch:          "branch1",
		Language:        "lang1",
		IsWrite:         true,
		Time:            CustomTime(now),
		Editor:          "editor1",
		OperatingSystem: "os1",
		Machine:         "machine1",
		UserAgent:       "ua1",
		Origin:          "origin1",
		OriginId:        "originid1",
		Lines:           10,
		LineNo:          5,
		CursorPos:       100,
	}
	h1.Hashed()

	// same values -> same hash
	h2 := *h1
	h2.Hashed()
	assert.Equal(t, h1.Hash, h2.Hash)

	// different UserID -> different hash
	h3 := *h1
	h3.UserID = "user2"
	h3.Hashed()
	assert.NotEqual(t, h1.Hash, h3.Hash)

	// different entity -> different hash
	h4 := *h1
	h4.Entity = "entity2"
	h4.Hashed()
	assert.NotEqual(t, h1.Hash, h4.Hash)

	// different type -> different hash
	h5 := *h1
	h5.Type = "domain"
	h5.Hashed()
	assert.NotEqual(t, h1.Hash, h5.Hash)

	// ...
	h6 := *h1
	h6.Category = "browsing"
	h6.Hashed()
	assert.NotEqual(t, h1.Hash, h6.Hash)

	// different time -> different hash
	h11 := *h1
	h11.Time = CustomTime(now.Add(1 * time.Second))
	h11.Hashed()
	assert.NotEqual(t, h1.Hash, h11.Hash)

	// different editor -> same hash
	h12 := *h1
	h12.Editor = "editor2"
	h12.Hashed()
	assert.Equal(t, h1.Hash, h12.Hash)

	// different OS -> same hash
	h13 := *h1
	h13.OperatingSystem = "os2"
	h13.Hashed()
	assert.Equal(t, h1.Hash, h13.Hash)

	// different machine -> same hash
	h14 := *h1
	h14.Machine = "machine2"
	h14.Hashed()
	assert.Equal(t, h1.Hash, h14.Hash)

	// different user agent -> same hash
	h15 := *h1
	h15.UserAgent = "ua2"
	h15.Hashed()
	assert.Equal(t, h1.Hash, h15.Hash)

	// different lines -> same hash
	h16 := *h1
	h16.Lines = 20
	h16.Hashed()
	assert.Equal(t, h1.Hash, h16.Hash)

	// different ID -> same hash
	h17 := *h1
	h17.ID = 123
	h17.Hashed()
	assert.Equal(t, h1.Hash, h17.Hash)

	// different origin -> same hash
	h18 := *h1
	h18.Origin = "origin2"
	h18.Hashed()
	assert.Equal(t, h1.Hash, h18.Hash)

	// different created date -> same hash
	h20 := *h1
	h20.CreatedAt = CustomTime(now.Add(1 * time.Hour))
	h20.Hashed()
	assert.Equal(t, h1.Hash, h20.Hash)

	// different origin -> same hash
	h21 := *h1
	h21.OriginId = "originid2"
	h21.Hashed()
	assert.Equal(t, h1.Hash, h21.Hash)
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

func TestHeartbeat_Unmarshal_IgnoreID(t *testing.T) {
	raw1 := "{\n    \"branch\":\"<<LAST_BRANCH>>\",\n    \"entity\":\"https://wakapi.dev\",\n    \"id\":\"f3647f89-e255-4dd1-8fcd-e20ba8f1709b\",\n    \"project\":\"<<LAST_PROJECT>>\",\n    \"time\":\"1728422364.044\",\n    \"type\":\"domain\",\n    \"userAgent\":\"Chrome/129.0.0.0 mac_x86-64 chrome-wakatime/4.0.6\"\n  }"
	raw2 := "{\n    \"branch\":\"<<LAST_BRANCH>>\",\n    \"entity\":\"https://wakapi.dev\",\n    \"id\":14,\n    \"project\":\"<<LAST_PROJECT>>\",\n    \"time\":\"1728422364.044\",\n    \"type\":\"domain\",\n    \"userAgent\":\"Chrome/129.0.0.0 mac_x86-64 chrome-wakatime/4.0.6\"\n  }"

	var parsed Heartbeat
	var err error

	// parse with string id
	err = json.Unmarshal([]byte(raw1), &parsed)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), parsed.ID)

	// parse with int id
	err = json.Unmarshal([]byte(raw2), &parsed)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), parsed.ID)

	parsed.ID = 14
	raw3, err := json.Marshal(parsed)
	assert.Nil(t, err)
	assert.NotContains(t, raw3, "\"id\":")
}
