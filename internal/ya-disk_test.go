package internal

import (
	"context"
	"log"
	"testing"
	"time"

	yadiskapi "finance-tg-bot/yaDiskApi"

	"github.com/stretchr/testify/assert"
)

var (
	client *yadiskapi.Client
	ctx    context.Context
)

func init() {
	var err error
	client, err = yadiskapi.NewClient(oAuth, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	ctx = context.Background()
}

func Test_DownloadFile_OK(t *testing.T) {
	err := DownloadFile(client, ctx)

	assert.NoError(t, err)
}

func Test_CreateBkp_OK(t *testing.T) {
	err := CreateBkp(client, ctx)

	assert.NoError(t, err)
}

// func Test_AddRecord_OK(t *testing.T) {
// 	err := AddRecord()

// 	assert.NoError(t, err)
// }

func Test_UploadFile_OK(t *testing.T) {
	err := UploadFile(client, ctx)

	assert.NoError(t, err)
}

func Test_runClear(t *testing.T) {
	assert.NoError(t, run())
}
