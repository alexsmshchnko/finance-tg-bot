package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const connStr string = "postgres://postgres:postgres@localhost:5433/base"

func Test_SayHello(t *testing.T) {
	db := NewDocumentStorage(connStr)
	err := db.SayHello(context.Background(), "vasya")

	assert.NoError(t, err)
}

func Test_GetCategories(t *testing.T) {
	db := NewDocumentStorage(connStr)
	res, err := db.GetCategories(context.Background(), "vasya")

	fmt.Println(res)

	assert.NoError(t, err)
}
