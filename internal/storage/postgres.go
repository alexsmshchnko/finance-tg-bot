package storage

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PGStorage struct {
	db sqlx.DB
}

func New(ctx context.Context, connString string) (*PGStorage, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", connString)
	if err != nil {
		return nil, err
	}
	return &PGStorage{db: *db}, nil
}

func (s *PGStorage) Close() {
	s.db.Close()
}
