package accountant

import (
	"context"
	"finance-tg-bot/internal/storage"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const connStr string = "postgres://postgres:postgres@localhost:5433/base"

func Test_SayHi(t *testing.T) {
	acc := NewAccountant(storage.NewDocumentStorage(connStr))
	err := acc.SayHi(context.Background(), "vasya")

	assert.NoError(t, err)
}

func Test_GetCats(t *testing.T) {
	acc := NewAccountant(storage.NewDocumentStorage(connStr))
	res, err := acc.GetCats(context.Background(), "vasya")

	fmt.Println(res)
	assert.NotEmpty(t, res)
	assert.NoError(t, err)
}
