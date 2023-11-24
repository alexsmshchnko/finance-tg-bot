package accountant

import (
	"context"
	"time"
)

type DocumentStorage interface {
	GetCategories(username string) ([]string, error)
	GetUserStatus(username string) (status bool, err error)
	PostDoc(time time.Time, category string, amount int, description string, msg_id string, client string) (err error)
	DeleteDoc(msg_id string, client string) (err error)
}

type Accountant struct {
	documents DocumentStorage
}

func NewAccountant(documentStorage DocumentStorage) *Accountant {
	return &Accountant{
		documents: documentStorage}
}

func (a *Accountant) GetCats(ctx context.Context, username string) (cats []string, err error) {
	cats, err = a.documents.GetCategories(username)
	return
}

func (a *Accountant) GetUserStatus(ctx context.Context, username string) (status bool, err error) {
	status, err = a.documents.GetUserStatus(username)
	return
}

func (a *Accountant) PostDoc(category string, amount int, description string, msg_id string, client string) (err error) {
	return a.documents.PostDoc(time.Now(), category, amount, description, msg_id, client)
}

func (a *Accountant) DeleteDoc(msg_id string, client string) (err error) {
	return a.documents.DeleteDoc(msg_id, client)
}
