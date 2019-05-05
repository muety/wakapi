package services

import (
	"database/sql"
	"fmt"

	"github.com/n1try/wakapi/models"
)

const TableUser = "user"

type UserService struct {
	Db *sql.DB
}

func (srv *UserService) GetUserById(userId string) (models.User, error) {
	q := fmt.Sprintf("SELECT user_id, api_key FROM %+s WHERE user_id = ?;", TableUser)
	u := models.User{}
	err := srv.Db.QueryRow(q, userId).Scan(&u.UserId, &u.ApiKey)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (srv *UserService) GetUserByKey(key string) (models.User, error) {
	q := fmt.Sprintf("SELECT user_id, api_key FROM %+s WHERE api_key = ?;", TableUser)
	var u models.User
	err := srv.Db.QueryRow(q, key).Scan(&u.UserId, &u.ApiKey)
	if err != nil {
		return u, err
	}
	return u, nil
}
