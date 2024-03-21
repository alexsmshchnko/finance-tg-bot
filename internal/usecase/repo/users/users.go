package users

import (
	"context"
	"database/sql"
	repPkg "finance-tg-bot/pkg/repository"
)

type User struct {
	repo repPkg.UserProvider
}

func New(repPkg repPkg.UserProvider) *User {
	return &User{repo: repPkg}
}

func (u *User) GetToken(ctx context.Context, username string) (string, error) {
	client, err := u.repo.GetUserInfo(ctx, username)
	if client.CloudToken.Valid {
		return client.CloudToken.String, nil
	}

	return "", err
}

func (u *User) GetStatus(ctx context.Context, username string) (status bool, err error) {
	client, err := u.repo.GetUserInfo(ctx, username)
	if err == sql.ErrNoRows || !client.IsActive.Bool {
		return false, nil
	}
	status = client.IsActive.Bool

	return
}
