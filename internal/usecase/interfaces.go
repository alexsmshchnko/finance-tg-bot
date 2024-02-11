package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"time"
)

type (
	Repo interface {
		GetCategories(username string) ([]string, error)
		GetSubCategories(username, trans_cat string) ([]string, error)
		PostDoc(time time.Time, category string, amount int, description string, msg_id string, direction int, client string) (err error)
		DeleteDoc(msg_id string, client string) (err error)
		GetUserStatus(username string) (status bool, err error)
		GetUserToken(username string) (token string, err error)
		ClearUserHistory(username string) (err error)
		Export(client string) (rslt []byte, err error)
		ImportDocs(data []byte, client string) (err error)
		GetStatement(p *entity.Report) (res string, err error)
	}
	Cloud interface {
		UploadFile(ctx context.Context, oAuth, filePath string) (err error)
		DownloadFile(ctx context.Context, oAuth, filePath string) (err error)
	}
)
