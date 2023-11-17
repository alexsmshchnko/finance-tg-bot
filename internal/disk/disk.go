package disk

import (
	"context"
	"fmt"
	"time"

	yaDisk "github.com/alexsmshchnko/ya-disk-api-client"
)

type CloudDisk struct {
	client   *yaDisk.Client
	userName string
	filePath string
}

func NewCloudDisk(oAuth, userName, filePath string) *CloudDisk {
	timeOut := 10 * time.Second
	client, _ := yaDisk.NewClient(oAuth, timeOut)
	return &CloudDisk{
		client:   client,
		userName: userName,
		filePath: filePath,
	}
}

// func CreateBkp(c *yaDisk.Client, ctx context.Context) (err error) {
// 	var sc int
// 	sc, err = c.MakeFolder(YA_DISK_BKP_PATH, ctx)
// 	if sc == 409 { //OK
// 		err = nil
// 	} else if err != nil {
// 		return
// 	}

// 	_, err = c.Copy(YA_DISK_FILE_PATH, YA_DISK_BKP_PATH+YA_DISK_FILE_NAME+time.Now().Format("060102150405")+YA_DISK_FILE_EXT, ctx)

// 	return
// }

func (c *CloudDisk) GetDiskInfo(ctx context.Context) (err error) {
	disk, sc, err := c.client.GetDiskInfo(ctx)

	fmt.Println(disk)
	fmt.Println(sc)
	return err
}

func (c *CloudDisk) GetFiles(ctx context.Context) (err error) {
	rsrc, sc, err := c.client.GetFiles(ctx)

	fmt.Println(rsrc)
	fmt.Println(sc)
	return err
}

func (c *CloudDisk) DownloadFile(ctx context.Context) (err error) {
	_, err = c.client.DownloadFile(c.filePath, "file-"+c.userName, ctx)
	return err
}

// func DownloadFile(c *yaDisk.Client, ctx context.Context) (err error) {
// 	_, err = c.DownloadFile(YA_DISK_FILE_PATH, YA_DISK_FILE_FULL_NAME, ctx)

// 	return
// }

// func UploadFile(c *yaDisk.Client, ctx context.Context) (err error) {
// 	//_, err = c.UploadFile(YA_DISK_FILE_PATH, "../"+YA_DISK_FILE_FULL_NAME, true, ctx)
// 	_, err = c.UploadFile(YA_DISK_FILE_PATH, YA_DISK_FILE_FULL_NAME, true, ctx)

// 	return
// }

// func initDownload(username string) (err error) {
// 	token, err := NewUser(username).GetUserToken()

// 	if err != nil {
// 		return err
// 	}

// 	client, err := yaDisk.NewClient(token, 10*time.Second)
// 	if err != nil {
// 		return err
// 	}

// 	err = DownloadFile(client, context.Background())
// 	if err != nil {
// 		return err
// 	}

// 	return
// }

// func SyncDiskFile(username string) (err error) {
// 	token, err := NewUser(username).GetUserToken()

// 	if err != nil {
// 		return err
// 	}

// 	client, err := yaDisk.NewClient(token, 10*time.Second)
// 	if err != nil {
// 		return err
// 	}

// 	// err = DownloadFile(client, context.Background())
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	return UploadFile(client, context.Background())
// }

// func run() error {
// 	client, err := yaDisk.NewClient(oAuth, 10*time.Second)

// 	if err != nil {
// 		return err
// 	}

// 	ctx := context.Background()

// 	disk, _, err := client.GetDiskInfo(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(disk)

// 	fr, _, err := client.GetFiles(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(fr)

// 	// link, _, err := client.GetDownloadLink(YA_DISK_FILE_PATH, ctx)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// fmt.Println(link)

// 	// resp, err := client.MkDir(YA_DISK_BKP_PATH, ctx)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// fmt.Println(resp)

// 	// err = CreateBkp(client, ctx)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return err
// 	// }

// 	// resp, err = client.CreateCopy(yadiskapi.YA_DISK_FILE_PATH,
// 	// 	yadiskapi.YA_DISK_BKP_PATH+yadiskapi.YA_DISK_FILE_NAME+time.Now().Format("060102150405")+yadiskapi.YA_DISK_FILE_EXT,
// 	// 	ctx)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	//fmt.Println(resp)

// 	// dt := time.Now().Format("060102150405")
// 	// erro := downloadFile("receipts"+dt+".xlsx", link)
// 	// if erro != nil {
// 	// 	return erro
// 	// }

// 	return nil
// }
