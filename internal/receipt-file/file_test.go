package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OpenFile(t *testing.T) {
	err := OpenFile("receipts231011125203.xlsx")
	assert.NoError(t, err)
}
