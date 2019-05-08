package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/n1try/wakapi/models"
)

const TableHeartbeat = "heartbeat"

type HeartbeatService struct {
	Db *sql.DB
}

func (srv *HeartbeatService) InsertBatch(heartbeats []*models.Heartbeat, user *models.User) error {
	qTpl := "INSERT INTO %+s (user, time, entity, type, category, is_write, project, branch, language, operating_system, editor) VALUES %+s;"
	qFill := ""
	vals := []interface{}{}

	for _, h := range heartbeats {
		qFill = qFill + "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
		vals = append(vals, user.UserId, h.Time.String(), h.Entity, h.Type, h.Category, h.IsWrite, h.Project, h.Branch, h.Language, h.OperatingSystem, h.Editor)
	}

	q := fmt.Sprintf(qTpl, TableHeartbeat, qFill[0:len(qFill)-1])
	stmt, _ := srv.Db.Prepare(q)
	result, err := stmt.Exec(vals...)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil || n != int64(len(heartbeats)) {
		return errors.New(fmt.Sprintf("Failed to insert %+v rows.", len(heartbeats)))
	}
	return nil
}

func (srv *HeartbeatService) GetAllFrom(date time.Time, user *models.User) ([]models.Heartbeat, error) {
	q := fmt.Sprintf("SELECT user, time, language, project, operating_system, editor FROM %+s WHERE time >= ? AND user = ?", TableHeartbeat)
	rows, err := srv.Db.Query(q, date.String(), user.UserId)
	defer rows.Close()
	if err != nil {
		return make([]models.Heartbeat, 0), err
	}

	var heartbeats []models.Heartbeat
	for rows.Next() {
		var h models.Heartbeat
		var language sql.NullString
		var project sql.NullString
		var operatingSystem sql.NullString
		var editor sql.NullString

		err := rows.Scan(&h.User, &h.Time, &language, &project, &operatingSystem, &editor)

		if language.Valid {
			h.Language = language.String
		}
		if project.Valid {
			h.Project = project.String
		}
		if operatingSystem.Valid {
			h.OperatingSystem = operatingSystem.String
		}
		if editor.Valid {
			h.Editor = editor.String
		}

		if err != nil {
			return make([]models.Heartbeat, 0), err
		}
		heartbeats = append(heartbeats, h)
	}
	return heartbeats, nil
}
