package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	yaDisk "github.com/alexsmshchnko/ya-disk-api-client"
)

const (
	timeOut = 10 * time.Second

	YA_DISK_APP_NAME  = "Финансовый бот"
	YA_DISK_FILE_NAME = "Семейный бюджет"
	YA_DISK_FILE_EXT  = ".xlsx"

	YA_DISK_FILE_FULL_NAME = YA_DISK_FILE_NAME + YA_DISK_FILE_EXT
	YA_DISK_BKP_PATH       = "disk:/Приложения/" + YA_DISK_APP_NAME + "/bkp/"
	YA_DISK_FILE_PATH      = "disk:/Приложения/" + YA_DISK_APP_NAME + "/" + YA_DISK_FILE_FULL_NAME
)

type Disk struct {
	diskAppPath string
}

func New() *Disk {
	return &Disk{}
}

func (c *Disk) getPaths(ctx context.Context, client *yaDisk.Client) (err error) {
	disk, _, err := client.GetDiskInfo(ctx)
	fmt.Printf("%#v\n", *disk)
	return
}

func (c *Disk) DownloadFile(ctx context.Context, oAuth, filePath string) (err error) {
	client, _ := yaDisk.NewClient(oAuth, timeOut)

	_, err = client.DownloadFile(YA_DISK_FILE_PATH, YA_DISK_FILE_FULL_NAME, ctx)
	return err
}

func getExportFileName(filePath string) string {
	return time.Now().Format("060102150405") + strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1]
}

func (c *Disk) UploadFile(ctx context.Context, oAuth, filePath string) (err error) {
	client, _ := yaDisk.NewClient(oAuth, timeOut)

	_, err = client.UploadFile(YA_DISK_BKP_PATH+getExportFileName(filePath), filePath, false, ctx)
	return err
}
