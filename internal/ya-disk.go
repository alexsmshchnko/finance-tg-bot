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

func DiskInfo() error {
	client, err := yadiskapi.NewClient(oAuth, 10*time.Second)

	if err != nil {
		return err
	}

	ctx := context.Background()

	disk, err := client.GetDiskInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Println(disk)

	fr, err := client.GetFiles(ctx, 100)
	if err != nil {
		return err
	}

	fmt.Println(fr)
	// link := fr.Embedded.Items[0].File
	// fmt.Println(link)

	return nil
}
