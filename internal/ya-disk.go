package internal

import (
	"context"
	yadiskapi "finance-tg-bot/yaDiskApi"
	"fmt"
	"os"
	"time"
)

var (
	oAuth string
)

func init() {

	if oAuth = os.Getenv("YA_DISK_AUTH_TOKEN"); oAuth == "" {
		panic(fmt.Errorf("failed to load env variable %s", "YA_DISK_AUTH_TOKEN"))
	}
}

func CreateBkp(c *yadiskapi.Client, ctx context.Context) (err error) {

	_, err = c.Copy(YA_DISK_FILE_PATH, YA_DISK_BKP_PATH+YA_DISK_FILE_NAME+time.Now().Format("060102150405")+YA_DISK_FILE_EXT,
		ctx)

	return
}

func run() error {
	client, err := yadiskapi.NewClient(oAuth, 10*time.Second)

	if err != nil {
		return err
	}

	ctx := context.Background()

	disk, _, err := client.GetDiskInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Println(disk)

	fr, _, err := client.GetFiles(ctx)
	if err != nil {
		return err
	}
	fmt.Println(fr)

	// link, _, err := client.GetDownloadLink(YA_DISK_FILE_PATH, ctx)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(link)

	// resp, err := client.MkDir(YA_DISK_BKP_PATH, ctx)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(resp)

	// err = CreateBkp(client, ctx)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// resp, err = client.CreateCopy(yadiskapi.YA_DISK_FILE_PATH,
	// 	yadiskapi.YA_DISK_BKP_PATH+yadiskapi.YA_DISK_FILE_NAME+time.Now().Format("060102150405")+yadiskapi.YA_DISK_FILE_EXT,
	// 	ctx)
	// if err != nil {
	// 	return err
	// }
	//fmt.Println(resp)

	// dt := time.Now().Format("060102150405")
	// erro := downloadFile("receipts"+dt+".xlsx", link)
	// if erro != nil {
	// 	return erro
	// }

	return nil
}
