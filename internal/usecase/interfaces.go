package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"log/slog"
	"time"
)

type (
	Repo interface {
		GetCategories(username string) ([]string, error)
		GetCats(ctx context.Context, username, limit string) (cat []entity.TransCatLimit, err error)
		EditCategory(tc entity.TransCatLimit, client string) (err error)
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
	Reporter interface {
		GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) (res string, err error)
	}
	Cloud interface {
		UploadFile(ctx context.Context, oAuth, filePath string) (err error)
		DownloadFile(ctx context.Context, oAuth, filePath string) (err error)
	}
)
