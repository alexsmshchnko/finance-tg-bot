package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetExpenseCategories(t *testing.T) {
	l := GetExpenseCategories()

	//fmt.Printf("%v\n", l)
	assert.NotEmpty(t, l)
}

// func Test_OpenFile(t *testing.T) {

// 	fmt.Println(a)

// 	err := OpenFile("receipts231011125203.xlsx")
// 	assert.NoError(t, err)
// }
