package accountant

import "context"

type DocumentStorage interface {
	GetCategories(username string) ([]string, error)
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
