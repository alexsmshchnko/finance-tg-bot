package users

import (
	"context"
	"database/sql"
	repPkg "finance-tg-bot/pkg/repository"
	"finance-tg-bot/pkg/ydb"
)

type User struct {
	ydb.DB
}

func New(db ydb.DB) *User {
	return &User{db}
}

func (db *User) GetToken(ctx context.Context, username string) (string, error) {
	client, err := repPkg.GetUserInfo(db.DB, ctx, username)
	if client.CloudToken.Valid {
		return client.CloudToken.String, nil
	}

	return "", err
}

func (db *User) GetStatus(ctx context.Context, username string) (status bool, err error) {
	client, err := repPkg.GetUserInfo(db.DB, ctx, username)
	if err == sql.ErrNoRows || !client.IsActive.Bool {
		return false, nil
	}
	status = client.IsActive.Bool

	return
}
