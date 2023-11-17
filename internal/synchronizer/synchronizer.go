package synchronizer

import "context"

type Downloader interface {
	DownloadFile(ctx context.Context) (err error)
}

type Uploader interface {
	UploadFile(ctx context.Context) (err error)
}

type Tokenizer interface {
	GetToken()
}

type Synchronizer struct {
	Downloader
	Uploader
}

func NewSynchronizer(downloader Downloader) *Synchronizer {
	return &Synchronizer{Downloader: downloader}
}
