package yadiskapi

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	YA_DISK_APP_NAME  = "Финансовый бот"
	YA_DISK_FILE_NAME = "receipts"
	YA_DISK_FILE_EXT  = ".xlsx"

	YA_DISK_FILE_FULL_NAME = YA_DISK_FILE_NAME + YA_DISK_FILE_EXT
	YA_DISK_BKP_PATH       = "disk:/Приложения/" + YA_DISK_APP_NAME + "/bkp/"
	YA_DISK_FILE_PATH      = "disk:/Приложения/" + YA_DISK_APP_NAME + "/" + YA_DISK_FILE_FULL_NAME
)

var logger = os.Stdout

type loggingRoundTripper struct {
	next http.RoundTripper
}

func (l loggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	fmt.Fprintf(logger, "[%s] %s %s\n", time.Now().Format(time.ANSIC), r.Method, r.URL)
	return l.next.RoundTrip(r)
}

var transport = loggingRoundTripper{
	next: http.DefaultTransport,
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	fmt.Fprintf(logger, "REDIRECT: %s", req.Response.Status)

	return nil
}
