package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_runClear(t *testing.T) {
	assert.NoError(t, run())
}

func Test_UploadFile_OK(t *testing.T) {
	assert.NoError(t, UploadFile("672241_v01_b.jpg", "https://uploader13g.disk.yandex.net:443/upload-target/20231012T201003"))
}
