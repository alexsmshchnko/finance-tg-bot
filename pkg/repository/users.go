package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
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
		r.ServiceDomain+"/users/"+username,
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetUserInfo http.NewRequestWithContext", "err", err)
		return
	}

	req.Header.Add("Authorization", "Basic "+r.AuthToken)
	client := &http.Client{}
	resp, err := client.Do(req)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error("Repository.GetUserInfo io.ReadAll", "err", err)
		return
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		r.Logger.Error("Repository.GetUserInfo json.Unmarshal", "err", err)
	}

	return
}
