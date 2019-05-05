package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/n1try/wakapi/models"
)

const TableHeartbeat = "heartbeat"

type HeartbeatService struct {
	Db *sql.DB
}

func (srv *HeartbeatService) InsertBatch(heartbeats []models.Heartbeat, user *models.User) error {
	qTpl := "INSERT INTO %+s (user, time, entity, type, category, is_write, project, branch, language) VALUES %+s;"
	qFill := ""
	vals := []interface{}{}

	for _, h := range heartbeats {
		qFill = "(?, ?, ?, ?, ?, ?, ?, ?, ?),"
		vals = append(vals, user.UserId, h.Time.String(), h.Entity, h.Type, h.Category, h.IsWrite, h.Project, h.Branch, h.Language)
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
