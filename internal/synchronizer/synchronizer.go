package synchronizer

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"finance-tg-bot/internal/local_storage"
	"fmt"
	"os"
	"time"
)

const (
	YA_DISK_APP_NAME  = "Финансовый бот"
	YA_DISK_FILE_NAME = "FamilyBudget"
	YA_DISK_FILE_EXT  = ".xlsx"

	YA_DISK_FILE_FULL_NAME = YA_DISK_FILE_NAME + YA_DISK_FILE_EXT
	YA_DISK_BKP_PATH       = "disk:/Приложения/" + YA_DISK_APP_NAME + "/bkp/"
	YA_DISK_FILE_PATH      = "disk:/Приложения/" + YA_DISK_APP_NAME + "/" + YA_DISK_FILE_FULL_NAME
)

type Loader interface {
	DownloadFile(ctx context.Context, oAuth, filePath string) (err error)
	UploadFile(ctx context.Context, oAuth, filePath string) (err error)
}

type DocExport struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	Description string    `json:"comment"`
	Direction   int       `json:"direction"`
}

type DB interface {
	GetUserToken(username string) (token string, err error)
	ClearUserHistory(username string) (err error)
	LoadDocs(time time.Time, category string, amount int, description string, direction int, client string) (err error)
	Export(client string) (rslt []byte, err error)
}

// type File interface {
// }

type Synchronizer struct {
	Loader
	DB
	// File
}

func New(loader Loader, db DB) *Synchronizer {
	return &Synchronizer{
		Loader: loader,
		DB:     db,
		// File: file,
	}
}

func (s *Synchronizer) PushToCloud(ctx context.Context, username string) (err error) {
	token, err := s.DB.GetUserToken(username)
	if err != nil {
		return
	}

	data, err := s.DB.Export(username)
	if err != nil {
		return err
	}

	var docs []DocExport
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
			fmt.Sprint(value.Amount), value.Description, fmt.Sprint(value.Direction)}
		err := writer.Write(rec)
		if err != nil {
			return err
		}
	}

	return s.Loader.UploadFile(ctx, token, file.Name())
}

func (s *Synchronizer) MigrateFromCloud(ctx context.Context, username string) (err error) {
	token, err := s.DB.GetUserToken(username)
	if err != nil {
		return
	}

	if err := s.DB.ClearUserHistory(username); err != nil {
		return err
	}

	err = s.Loader.DownloadFile(ctx, token, YA_DISK_FILE_PATH)
	if err != nil {
		return
	}

	rowsToSync, err := local_storage.GetRowsToSync(YA_DISK_FILE_FULL_NAME)
	if err != nil {
		return
	}

	for _, v := range rowsToSync {
		if err = s.LoadDocs(v.Time, v.Category, v.Amount, v.Description, v.Direction, username); err != nil {
			return err
		}
	}

	return
}
