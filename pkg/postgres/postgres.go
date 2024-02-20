package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Postgres struct {
	sqlx.DB
}

func New(ctx context.Context, connString string) (*Postgres, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", connString)
	if err != nil {
		return nil, err
	}
	return &Postgres{*db}, nil
}

func (s *Postgres) Close() {
	s.Close()
}
