package yadiskapi

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	AUTH_TOKEN = os.Getenv("YA_DISK_AUTH_TOKEN")
}

var (
	AUTH_TOKEN string
)

const (
	CLIENT_TIMEOUT = 10 * time.Second
)

func Test_NewClient_AutMissingErr(t *testing.T) {
	client, err := NewClient("", CLIENT_TIMEOUT)

	assert.Error(t, err)
	assert.EqualError(t, err, "oAuth is missing")
	assert.Empty(t, client)
}

func Test_NewClient_NoErr(t *testing.T) {
	client, err := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)

	assert.NoError(t, err)
	assert.NotEmpty(t, client)
}

func Test_GetDiskInfo_NotAuthorizedErr(t *testing.T) {
	client, _ := NewClient("errAuthToken", CLIENT_TIMEOUT)
	resp, statusCode, err := client.GetDiskInfo(context.Background())

	assert.Error(t, err)
	assert.Equal(t, 401, statusCode)
	assert.Empty(t, resp)
}

func Test_GetDiskInfo_OK(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	resp, statusCode, err := client.GetDiskInfo(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.NotEmpty(t, resp)
}

func Test_MkDir_OK(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	statusCode, err := client.MakeFolder("disk:/Приложения/Финансовый бот/t"+time.Now().Format("060102150405"), context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 201, statusCode)
}

func Test_MkDir_PathAlreadyExists(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	statusCode, err := client.MakeFolder("disk:/Приложения/Финансовый бот/bkp", context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 409, statusCode)
}

func Test_GetDownloadLink_WrongPath(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	href, statusCode, err := client.GetDownloadLink("errPath", context.Background())

	assert.Error(t, err)
	assert.Equal(t, 404, statusCode)
	assert.Empty(t, href)
}

func Test_GetDownloadLink_OK(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	href, statusCode, err := client.GetDownloadLink("disk:/Приложения/Финансовый бот/receipts.xlsx", context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.NotEmpty(t, href)
}

func Test_GetUploadLink_AlreadyExistsErr(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	link, statusCode, err := client.GetUploadLink("disk:/Приложения/Финансовый бот/receipts.xlsx", false, context.Background())

	assert.Error(t, err)
	assert.Equal(t, 409, statusCode)
	assert.Empty(t, link)
}

func Test_GetUploadLink_OK(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	link, statusCode, err := client.GetUploadLink("disk:/Приложения/Финансовый бот/receipts.xlsx", true, context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.NotEmpty(t, link)
}

func Test_UploadFile_OK(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	link, statusCode, err := client.UploadFile("disk:/Приложения/Финансовый бот/receipts.xlsx", true, context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.NotEmpty(t, link)
}

func Test_GetOperation_WrongOperation(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	status, statusCode, err := client.GetOperation("someOperation", context.Background())

	assert.Error(t, err)
	assert.Equal(t, 404, statusCode)
	assert.Empty(t, status)
}

func Test_GetOperation_OK(t *testing.T) {
	client, _ := NewClient(AUTH_TOKEN, CLIENT_TIMEOUT)
	status, statusCode, err := client.GetOperation("22c2aed3341468d45fbcc878451422d6d1c841a8c4700e045a02e2776f371d9d", context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.NotEmpty(t, status)
}
