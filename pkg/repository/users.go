package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

type UserProvider interface {
	GetUserInfo(ctx context.Context, username string) (*DBClient, error)
}

type DBClient struct {
	ID         sql.NullInt64  `db:"id"`
	Username   sql.NullString `db:"username"`
	FirstLogin sql.NullTime   `db:"first_login_date"`
	IsActive   sql.NullBool   `db:"is_active"`
	CloudName  sql.NullString `db:"external_system_name"`
	CloudToken sql.NullString `db:"external_system_token"`
}

func (r *Repository) GetUserInfo(ctx context.Context, username string) (res *DBClient, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		r.serviceURL.JoinPath("users", username).String(),
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetUserInfo http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetUserInfo client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.GetUserInfo response", "StatusCode", resp.StatusCode)
		return
	}

	err = json.NewDecoder(req.Body).Decode(&res)
	if err != nil {
		r.Logger.Error("json.NewDecoder(req.Body).Decode(&res)", "err", err)
	}

	return
}
