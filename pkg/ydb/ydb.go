package ydb

import (
	"context"
	"database/sql"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	yc "github.com/ydb-platform/ydb-go-yc"
)

type DB struct {
	*sql.DB
}

type Ydb struct {
	*ydb.Driver
}

func NewNative(ctx context.Context, dsn, saPath string) (*Ydb, error) {
	nativeDriver, err := ydb.Open(ctx, dsn, yc.WithInternalCA(),
		yc.WithServiceAccountKeyFileCredentials(saPath), // auth from service account key file
		//yc.WithMetadataCredentials(ctx), // auth inside cloud (virual machine or yandex function)
	)
	return &Ydb{nativeDriver}, err
}

func New(ctx context.Context, dsn, saPath string) (*DB, error) {
	nativeDriver, err := ydb.Open(ctx, dsn, yc.WithInternalCA(),
		yc.WithServiceAccountKeyFileCredentials(saPath), // auth from service account key file
		//yc.WithMetadataCredentials(ctx), // auth inside cloud (virual machine or yandex function)
	)
	if err != nil {
		return nil, err
	}

	connector, err := ydb.Connector(nativeDriver, ydb.WithAutoDeclare())
	// See ydb.ConnectorOption's for configure connector https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#ConnectorOption

	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	db.SetConnMaxIdleTime(time.Second) // workaround for background keep-aliving of YDB sessions
	return &DB{db}, nil
}
