package repo

import (
	"database/sql"
	"time"
)

type DBClient struct {
	ID         *int       `db:"id"`
	Username   *string    `db:"username"`
	FirstLogin *time.Time `db:"first_login_date"`
	IsActive   *bool      `db:"is_active"`
	CloudName  *string    `db:"external_system_name"`
	CloudToken *string    `db:"external_system_token"`
}

func (s *Repo) getUserInfo(username string) (*DBClient, error) {
	var client DBClient
	err := s.Db.Get(&client, "select * from client where username = $1", username)
	return &client, err
}

func (s *Repo) GetUserToken(username string) (token string, err error) {
	client, err := s.getUserInfo(username)
	if client.CloudToken != nil {
		token = *client.CloudToken
	}

	return
}

func (s *Repo) GetUserStatus(username string) (status bool, err error) {
	client, err := s.getUserInfo(username)
	if err == sql.ErrNoRows || client.IsActive == nil {
		return false, nil
	}
	status = *client.IsActive

	return
}
