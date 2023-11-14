package accountant

import "context"

type DocumentStorage interface {
	SayHello(ctx context.Context, username string) error
	GetCategories(ctx context.Context, username string) ([]string, error)
}

type Accountant struct {
	documents DocumentStorage
}

func NewAccountant(documentStorage DocumentStorage) *Accountant {
	return &Accountant{
		documents: documentStorage}
}

func (a *Accountant) GetCats(ctx context.Context, username string) (cats []string, err error) {
	cats, err = a.documents.GetCategories(ctx, username)
	return
}

func (a *Accountant) SayHi(ctx context.Context, username string) (err error) {
	err = a.documents.SayHello(ctx, username)
	return
}
