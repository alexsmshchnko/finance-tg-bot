package usecase

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"finance-tg-bot/internal/entity"
	"finance-tg-bot/internal/local_storage"
	"fmt"
	"os"
)

const (
	YA_DISK_APP_NAME  = "Финансовый бот"
	YA_DISK_FILE_NAME = "Семейный бюджет"
	YA_DISK_FILE_EXT  = ".xlsx"

	YA_DISK_FILE_FULL_NAME = YA_DISK_FILE_NAME + YA_DISK_FILE_EXT
	YA_DISK_BKP_PATH       = "disk:/Приложения/" + YA_DISK_APP_NAME + "/bkp/"
	YA_DISK_FILE_PATH      = "disk:/Приложения/" + YA_DISK_APP_NAME + "/" + YA_DISK_FILE_FULL_NAME
)

func (a *Accountant) PushToCloud(ctx context.Context, username string) (err error) {
	token, err := a.repo.GetUserToken(username)
	if err != nil {
		return
	}

	data, err := a.repo.Export(username)
	if err != nil {
		return err
	}

	var docs []entity.DocumentExport
	err = json.Unmarshal(data, &docs)
	if err != nil {
		return err
	}

	file, err := os.Create("export_" + username + ".csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range docs {
		rec := []string{value.Time.UTC().Format("02.01.2006"), value.Category,
			fmt.Sprint(value.Amount), value.Description, value.Direction}
		err := writer.Write(rec)
		if err != nil {
			return err
		}
	}

	return a.sync.UploadFile(ctx, token, file.Name())
}

func (a *Accountant) MigrateFromCloud(ctx context.Context, username string) (err error) {
	err = a.PushToCloud(ctx, username)
	if err != nil {
		return
	}

	token, err := a.repo.GetUserToken(username)
	if err != nil {
		return
	}

	if err := a.repo.ClearUserHistory(username); err != nil {
		return err
	}

	err = a.sync.DownloadFile(ctx, token, YA_DISK_FILE_PATH)
	if err != nil {
		return
	}

	rowsToSync, err := local_storage.GetRowsToSync(YA_DISK_FILE_FULL_NAME)
	if err != nil {
		return
	}

	data, err := json.Marshal(rowsToSync)
	if err != nil {
		return
	}
	return a.repo.ImportDocs(data, username)

}
