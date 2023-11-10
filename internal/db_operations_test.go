package internal

import (
	_ "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testUser string = "testovich"

func Test_UserToken_OK(t *testing.T) {
	tkn, err := NewUser(testUser).GetUserToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, tkn)
}

func Test_UserExpense_OK(t *testing.T) {
	exp := NewFinRec("здоровье", 5000, "psyho", "0")

	err := NewUser(testUser).NewExpense(exp)

	assert.NoError(t, err)
}
