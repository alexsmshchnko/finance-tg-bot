package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"time"
)

type Accountant struct {
	repo     Repo
	reporter Reporter
	sync     Cloud
}

func New(d Repo, r Reporter, s Cloud) *Accountant {
	return &Accountant{
		repo:     d,
		reporter: r,
		sync:     s,
	}
}

func (a *Accountant) GetCats(ctx context.Context, username string) (cats []string, err error) {
	cats, err = a.repo.GetCategories(username)
	return
}

func (a *Accountant) GetCatsLimit(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error) {
	cats, err = a.repo.GetCats(ctx, username, limit)
	return
}

func (a *Accountant) EditCats(tc entity.TransCatLimit, client string) (err error) {
	err = a.repo.EditCategory(tc, client)
	return
}

func (a *Accountant) GetSubCats(ctx context.Context, username, trans_cat string) (cats []string, err error) {
	cats, err = a.repo.GetSubCategories(username, trans_cat)
	return
}

func (a *Accountant) GetUserStatus(ctx context.Context, username string) (status bool, err error) {
	status, err = a.repo.GetUserStatus(username)
	return
}

func (a *Accountant) PostDoc(category string, amount int, description string, msg_id string, direction int, client string) (err error) {
	return a.repo.PostDoc(time.Now(), category, amount, description, msg_id, direction, client)
}

func (a *Accountant) DeleteDoc(msg_id string, client string) (err error) {
	return a.repo.DeleteDoc(msg_id, client)
}

func (a *Accountant) GetStatement(p map[string]string) (res string, err error) {
	return a.reporter.GetStatementTotals(context.Background(), nil, p)
}
