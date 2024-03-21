package ydb

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	yc "github.com/ydb-platform/ydb-go-yc"
)

type Ydb struct {
	*ydb.Driver
}

func NewNative(ctx context.Context, dsn, saPath string) (*Ydb, error) {
	var opt ydb.Option
	if saPath == "" {
		// auth inside cloud (virual machine or yandex function)
		opt = yc.WithMetadataCredentials()
	} else {
		// auth from service account key file
		opt = yc.WithServiceAccountKeyFileCredentials(saPath)
	}
	nativeDriver, err := ydb.Open(ctx, dsn, yc.WithInternalCA(), opt)
	return &Ydb{nativeDriver}, err
}
