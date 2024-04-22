package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"log/slog"
)

type (
	Repo interface {
		GetCategories(ctx context.Context, username, limit string) (cat []entity.TransCatLimit, err error)
		EditCategory(ctx context.Context, tc entity.TransCatLimit, client string) (err error) //TODO include client
		GetSubCategories(ctx context.Context, username, trans_cat string) ([]string, error)
		Export(client string) (rslt []byte, err error)
		ImportDocs(data []byte, client string) (err error)
		PostDocument(ctx context.Context, doc *entity.Document) (err error)
		DeleteDocument(ctx context.Context, doc *entity.Document) (err error)
	}
	User interface {
		GetStatus(ctx context.Context, username string) (id int, status bool, err error)
		GetToken(ctx context.Context, username string) (token string, err error)
	}
	Reporter interface {
		GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) (res string, err error)
		GetUserStats(ctx context.Context, user_id int) (stats entity.UserStats, err error)
	}
	Cloud interface {
		UploadFile(ctx context.Context, oAuth, filePath string) (err error)
		DownloadFile(ctx context.Context, oAuth, filePath string) (err error)
	}
)
