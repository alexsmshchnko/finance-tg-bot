package yadiskapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func download(filepath string, url string) (statusCode int, err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check server response
	statusCode = resp.StatusCode
	if statusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %d", statusCode)
		return
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	return
}

func upload(filepath string, url string) (statusCode int, err error) {
	data, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer data.Close()

	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return
	}
	//req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	res, err := client.Do(req)
	statusCode = res.StatusCode
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}
