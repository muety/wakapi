package models

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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
	assert.Equal(t, UnknownSummaryKey, sut.GetKey(SummaryAiModel))
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
		AIModel:         "claude-3-5-sonnet",
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

	// different ai model -> same hash
	h22 := *h1
	h22.AIModel = "gpt-4o"
	h22.Hashed()
	assert.Equal(t, h1.Hash, h22.Hash)

	// different ai cached input tokens -> same hash
	h23 := *h1
	h23.AICachedInputTokens = 4096
	h23.Hashed()
	assert.Equal(t, h1.Hash, h23.Hash)
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

func TestHeartbeat_AICachedInputTokens_JSON(t *testing.T) {
	// JSON unmarshaling with ai_cached_input_tokens populated
	raw := `{"entity":"test.go","type":"file","category":"ai coding","project":"wakapi","ai_input_tokens":100,"ai_output_tokens":50,"ai_cached_input_tokens":2500}`
	var hb Heartbeat
	err := json.Unmarshal([]byte(raw), &hb)
	require.NoError(t, err)
	assert.Equal(t, 100, hb.AIInputTokens)
	assert.Equal(t, 50, hb.AIOutputTokens)
	assert.Equal(t, 2500, hb.AICachedInputTokens)

	// JSON unmarshaling when ai_cached_input_tokens is omitted
	rawOmitted := `{"entity":"test.go","type":"file","category":"ai coding","project":"wakapi","ai_input_tokens":100}`
	var hbOmitted Heartbeat
	err = json.Unmarshal([]byte(rawOmitted), &hbOmitted)
	require.NoError(t, err)
	assert.Equal(t, 100, hbOmitted.AIInputTokens)
	assert.Equal(t, 0, hbOmitted.AICachedInputTokens)

	// JSON unmarshaling with zero
	rawZero := `{"entity":"test.go","type":"file","category":"ai coding","project":"wakapi","ai_cached_input_tokens":0}`
	var hbZero Heartbeat
	err = json.Unmarshal([]byte(rawZero), &hbZero)
	require.NoError(t, err)
	assert.Equal(t, 0, hbZero.AICachedInputTokens)

	// JSON marshaling includes field when non-zero
	marshaled, err := json.Marshal(&hb)
	require.NoError(t, err)
	assert.Contains(t, string(marshaled), `"ai_cached_input_tokens":2500`)

	// JSON marshaling omits field when zero due to omitempty
	marshaledOmitted, err := json.Marshal(&hbOmitted)
	require.NoError(t, err)
	assert.NotContains(t, string(marshaledOmitted), "ai_cached_input_tokens")
}

func TestHeartbeat_AICachedInputTokens_MigrationAndPersistence(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wakapi_migration_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	sqlDb, err := db.DB()
	require.NoError(t, err)
	defer sqlDb.Close()

	// Simulate existing Wakapi schema without ai_cached_input_tokens column
	createLegacyHeartbeatsTable := `CREATE TABLE heartbeats (
		id integer PRIMARY KEY AUTOINCREMENT,
		user_id text NOT NULL,
		entity text NOT NULL,
		type text,
		category text,
		project text,
		branch text,
		language text,
		is_write numeric,
		editor text,
		operating_system text,
		machine text,
		ai_model text,
		user_agent varchar(255),
		time integer NOT NULL,
		hash varchar(17),
		origin varchar(255),
		origin_id varchar(255),
		created_at integer NOT NULL,
		lines integer,
		line_no integer,
		cursor_pos integer,
		line_deletions integer,
		line_additions integer,
		project_root_count integer,
		ai_line_changes integer,
		ai_session text,
		ai_input_tokens integer,
		ai_output_tokens integer,
		ai_prompt_length integer,
		human_line_changes integer
	);`
	require.NoError(t, db.Exec(createLegacyHeartbeatsTable).Error)

	now := time.Now()
	legacyTime := CustomTime(now)
	// Insert an existing row from an older Wakapi version
	insertLegacyRow := `INSERT INTO heartbeats (user_id, entity, time, created_at, ai_input_tokens, ai_output_tokens)
		VALUES ('user_old', 'main.go', ?, ?, 100, 50);`
	require.NoError(t, db.Exec(insertLegacyRow, legacyTime, legacyTime).Error)

	// Run AutoMigrate on the modern Heartbeat model (with AICachedInputTokens)
	require.NoError(t, db.AutoMigrate(&Heartbeat{}))

	// Verify table schema contains ai_cached_input_tokens column
	type TableColumn struct {
		Name string `gorm:"column:name"`
	}
	var columns []TableColumn
	require.NoError(t, db.Raw("PRAGMA table_info(heartbeats)").Scan(&columns).Error)
	columnNames := make([]string, len(columns))
	for i, c := range columns {
		columnNames[i] = c.Name
	}
	assert.Contains(t, columnNames, "ai_cached_input_tokens")

	// Verify the legacy row reads back with AICachedInputTokens = 0
	var oldHb Heartbeat
	require.NoError(t, db.Where("user_id = ?", "user_old").First(&oldHb).Error)
	assert.Equal(t, 100, oldHb.AIInputTokens)
	assert.Equal(t, 50, oldHb.AIOutputTokens)
	assert.Equal(t, 0, oldHb.AICachedInputTokens)

	// Insert a new Heartbeat with AICachedInputTokens populated
	newHb := &Heartbeat{
		UserID:              "user_new",
		Entity:              "worker.go",
		Time:                CustomTime(now.Add(time.Minute)),
		CreatedAt:           CustomTime(now),
		AIInputTokens:       300,
		AIOutputTokens:      150,
		AICachedInputTokens: 7500,
	}
	newHb.Hashed()
	require.NoError(t, db.Create(newHb).Error)

	// Verify the new row persists and reads back accurately
	var retrievedHb Heartbeat
	require.NoError(t, db.Where("user_id = ?", "user_new").First(&retrievedHb).Error)
	assert.Equal(t, 300, retrievedHb.AIInputTokens)
	assert.Equal(t, 150, retrievedHb.AIOutputTokens)
	assert.Equal(t, 7500, retrievedHb.AICachedInputTokens)
}
