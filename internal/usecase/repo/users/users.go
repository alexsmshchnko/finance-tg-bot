package users

import (
	"context"
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
	if err == nil && client != nil {
		status = client.IsActive.Bool
	}
	return status, nil
}
