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

func (a *Accountant) GetUserStatus(ctx context.Context, username string) (status bool, err error) {
	a.log.Debug("GetUserStatus request", "username", username)
	status, err = a.user.GetStatus(ctx, username)
	if err != nil {
		a.log.Error("user.GetStatus", "err", err)
	}
	a.log.Debug("GetUserStatus response", "username", username, "status", status)
	return
}

func (a *Accountant) PostDoc(ctx context.Context, category string, amount int, description string, msg_id string, direction int, client string) (err error) {
	a.log.Debug("PostDoc", "client", client, "category", category, "msg_id", msg_id)
	err = a.repo.PostDocument(ctx,
		&entity.Document{
			Category:    category,
			Amount:      int64(amount),
			Description: description,
			MsgID:       msg_id,
			ChatID:      "",
			ClientID:    client,
			Direction:   int16(direction),
		},
	)
	if err != nil {
		a.log.Error("nrepo.PostDocument", "err", err)
	}

	return
}

func (a *Accountant) DeleteDoc(msg_id string, client string) (err error) {
	a.log.Debug("DeleteDoc", "client", client, "msg_id", msg_id)
	err = a.repo.DeleteDocument(context.Background(),
		&entity.Document{
			MsgID:    msg_id,
			ChatID:   "",
			ClientID: client,
		},
	)
	if err != nil {
		a.log.Error("nrepo.DeleteDocument", "err", err)
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
