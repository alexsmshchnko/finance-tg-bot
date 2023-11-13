package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_runClear(t *testing.T) {
	assert.NoError(t, run())
}

func Test_botUser(t *testing.T) {
	l := BotUsers[BOT_ADMIN]

	l.getUserDiskToken()
}
