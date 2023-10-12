package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_runClear(t *testing.T) {
	assert.NoError(t, run())
}
