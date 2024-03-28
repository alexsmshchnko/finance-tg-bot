package repository

import (
	"context"
	"database/sql"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
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

func (r *Repository) GetUserInfo(ctx context.Context, username string) (*DBClient, error) {
	var (
		client DBClient
		qError error = sql.ErrNoRows
	)

	err := r.Ydb.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			`DECLARE $uname AS String;
			SELECT id, username, first_login_date,
				   is_active,
				   external_system_name, external_system_token
			  FROM client
			 WHERE username = $uname;`,
			table.NewQueryParameters(table.ValueParam("$uname", types.BytesValueFromString(username))),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			qError = nil
			err = res.ScanNamed(
				named.Required("id", &client.ID),
				named.Optional("username", &client.Username),
				named.Optional("first_login_date", &client.FirstLogin),
				named.OptionalWithDefault("is_active", &client.IsActive),
				named.Optional("external_system_name", &client.CloudName),
				named.Optional("external_system_token", &client.CloudToken),
			)
			if err != nil {
				return err
			}
		}
		return res.Err() // for driver retry if not nil
	})

	if err != nil {
		r.Logger.Error("Repository.GetUserInfo r.Ydb.Table().Do", "err", err)
		return nil, err
	}

	return &client, qError
}
