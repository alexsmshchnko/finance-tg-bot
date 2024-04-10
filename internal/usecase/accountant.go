package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"log/slog"
)

type Accountant struct {
	repo     Repo
	user     User
	reporter Reporter
	sync     Cloud
	log      *slog.Logger
}

func New(d Repo, u User, r Reporter, s Cloud, l *slog.Logger) *Accountant {
	return &Accountant{
		repo:     d,
		user:     u,
		reporter: r,
		sync:     s,
		log:      l,
	}
}

func (a *Accountant) GetCatsLimit(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error) {
	a.log.Debug("GetCatsLimit", "username", username, "limit", limit)
	cats, err = a.repo.GetCategories(ctx, username, limit)
	if err != nil {
		a.log.Error("repo.GetCategories", "err", err)
	}
	return
}

func (a *Accountant) EditCats(ctx context.Context, tc entity.TransCatLimit, client string) (err error) {
	a.log.Debug("EditCats", "client", client)
	err = a.repo.EditCategory(ctx, tc, client)
	if err != nil {
		a.log.Error("repo.EditCategory", "err", err)
	}
	return
}

func (a *Accountant) GetSubCats(ctx context.Context, username, trans_cat string) (cats []string, err error) {
	a.log.Debug("GetSubCats", "username", username, "trans_cat", trans_cat)
	cats, err = a.repo.GetSubCategories(ctx, username, trans_cat)
	if err != nil {
		a.log.Error("repo.GetSubCategories", "err", err)
	}
	return
}

func (a *Accountant) GetUserStatus(ctx context.Context, username string) (id int, status bool, err error) {
	a.log.Debug("GetUserStatus request", "username", username)
	id, status, err = a.user.GetStatus(ctx, username)
	if err != nil {
		a.log.Error("user.GetStatus", "err", err)
	}
	a.log.Debug("GetUserStatus response", "username", username, "status", status)
	return
}

func (a *Accountant) PostDoc(ctx context.Context, doc *entity.Document) (err error) {
	a.log.Debug("PostDoc", "client", doc.ClientID, "category", doc.Category, "msg_id", doc.MsgID)
	err = a.repo.PostDocument(ctx, doc)
	if err != nil {
		a.log.Error("nrepo.PostDocument", "err", err)
	}
	return
}

func (a *Accountant) DeleteDoc(chat_id, msg_id, client string) (err error) {
	a.log.Debug("DeleteDoc", "chat_id", chat_id, "msg_id", msg_id, "client", client)
	err = a.repo.DeleteDocument(context.Background(),
		&entity.Document{
			MsgID:    msg_id,
			ChatID:   chat_id,
			ClientID: client,
		},
	)
	if err != nil {
		a.log.Error("repo.DeleteDocument", "err", err)
	}
	return
}

func (a *Accountant) GetStatement(p map[string]string) (res string, err error) {
	a.log.Debug("GetStatementTotals", "p", p)
	res, err = a.reporter.GetStatementTotals(context.Background(), a.log, p)
	if err != nil {
		a.log.Error("reporter.GetStatementTotals", "err", err)
	}
	return
}
