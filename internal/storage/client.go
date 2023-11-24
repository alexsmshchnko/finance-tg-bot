package storage

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PGStorage struct {
	db sqlx.DB
}

func NewPGStorage(ctx context.Context, connString string) *PGStorage {
	db, err := sqlx.ConnectContext(ctx, "postgres", connString)
	if err != nil {
		log.Fatal(err)
	}
	return &PGStorage{db: *db}
}

type DBClient struct {
	ID         int       `db:"id"`
	Username   string    `db:"username"`
	FirstLogin time.Time `db:"first_login_date"`
	IsActive   bool      `db:"is_active"`
	CloudName  string    `db:"external_system_name"`
	CloudToken string    `db:"external_system_token"`
}

func (s *PGStorage) getUserInfo(username string) (*DBClient, error) {
	var client DBClient
	err := s.db.Get(&client, "select * from client where username = $1", username)
	return &client, err
}

func (s *PGStorage) GetUserToken(username string) (token string, err error) {
	client, err := s.getUserInfo(username)

	token = client.CloudToken

	return
}

func (s *PGStorage) GetUserStatus(username string) (status bool, err error) {
	client, err := s.getUserInfo(username)

	status = client.IsActive

	return
}
