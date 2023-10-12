package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func UploadFile(filepath string, url string) error {
	data, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer data.Close()

	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
