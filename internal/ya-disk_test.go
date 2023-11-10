package internal

import (
	"context"
	"log"
	"testing"
	"time"

	yadiskapi "github.com/alexsmshchnko/ya-disk-api-client"

	"github.com/stretchr/testify/assert"
)

var (
	client *yadiskapi.Client
	ctx    context.Context
)

func init_test() {
	var err error
	client, err = yadiskapi.NewClient(oAuth, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	ctx = context.Background()
}

func Test_DownloadFile_OK(t *testing.T) {
	init_test()
	err := DownloadFile(client, ctx)

	assert.NoError(t, err)
}

func Test_CreateBkp_OK(t *testing.T) {
	init_test()
	err := CreateBkp(client, ctx)

	assert.NoError(t, err)
}

// func Test_AddRecord_OK(t *testing.T) {
// 	err := AddRecord()

// 	assert.NoError(t, err)
// }

func Test_UploadFile_OK(t *testing.T) {
	init_test()
	err := UploadFile(client, ctx)

	assert.NoError(t, err)
}

func Test_runClear(t *testing.T) {
	init_test()
	assert.NoError(t, run())
}

func Test_SyncDiskFile_OK(t *testing.T) {
	err := SyncDiskFile("testovich")

	assert.NoError(t, err)
}
