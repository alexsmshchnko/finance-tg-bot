package yadiskapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://cloud-api.yandex.net/v1/disk/"

type Client struct {
	oAuth   string
	baseURl string
	client  *http.Client
}

func NewClient(oAuth string, timeout time.Duration) (*Client, error) {
	if timeout == 0 {
		return nil, errors.New("timeout can't be zero")
	}

	return &Client{
		oAuth:   oAuth,
		baseURl: baseURL,
		client: &http.Client{
			Timeout:       timeout,
			Transport:     transport,
			CheckRedirect: checkRedirect,
		},
	}, nil
}

func (c *Client) sendRequest(req *http.Request, data interface{}) error {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "OAuth "+c.oAuth)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		errorResponse := ErrorResponse{}
		if err = json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("%sstatus code: %d\n", errorResponse.Info(), resp.StatusCode)
		}

		return fmt.Errorf("unknown error, status code: %d\n", resp.StatusCode)
	}

	json.NewDecoder(resp.Body).Decode(&data)

	return nil
}

func (c *Client) GetDiskInfo(ctx context.Context) (*Disk, error) {
	req, err := http.NewRequest("GET", c.baseURl, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	disk := Disk{}

	if err = c.sendRequest(req, &disk); err != nil {
		return nil, err
	}

	return &disk, nil

}

func (c *Client) GetFiles(ctx context.Context, limit int) (*ResourceList, error) {
	req, err := http.NewRequest("GET", c.baseURl+"resources?path=app:/", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	req = req.WithContext(ctx)

	filesResourceList := ResourceList{}

	if err = c.sendRequest(req, &filesResourceList); err != nil {
		return nil, err
	}

	return &filesResourceList, nil
}
