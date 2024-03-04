package repository

import (
	"context"
	"database/sql"
	"finance-tg-bot/pkg/ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type DBClient struct {
	ID         sql.NullInt64  `db:"id"`
	Username   sql.NullString `db:"username"`
	FirstLogin sql.NullTime   `db:"first_login_date"`
	IsActive   sql.NullBool   `db:"is_active"`
	CloudName  sql.NullString `db:"external_system_name"`
	CloudToken sql.NullString `db:"external_system_token"`
}

func GetUserInfo(db ydb.DB, ctx context.Context, username string) (*DBClient, error) {
	var client DBClient

	query := `DECLARE $uname AS String;
			SELECT
				id,
				username,
				first_login_date,
				is_active,
				external_system_name,
				external_system_token
			FROM
				client
			WHERE
				username = $uname;`

	err := db.QueryRowContext(ctx, query, table.NewQueryParameters(table.ValueParam("$uname", types.BytesValueFromString(username)))).
		Scan(&client.ID, &client.Username, &client.FirstLogin, &client.IsActive, &client.CloudName, &client.CloudToken)

	return &client, err
}