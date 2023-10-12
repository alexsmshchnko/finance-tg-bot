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

const (
	baseURL     = "https://cloud-api.yandex.net/v1/disk/"
	resourePath = "resources/"
)

type Client struct {
	oAuth   string
	baseURl string
	client  *http.Client
}

func NewClient(oAuth string, timeout time.Duration) (client *Client, err error) {
	if timeout == 0 {
		err = errors.New("timeout can't be zero")
		return
	} else if oAuth == "" {
		err = errors.New("oAuth is missing")
		return
	}

	client = &Client{
		oAuth:   oAuth,
		baseURl: baseURL,
		client: &http.Client{
			Timeout:       timeout,
			Transport:     transport,
			CheckRedirect: checkRedirect,
		},
	}

	return
}

func (c *Client) sendReq(request *http.Request, response interface{}) (statusCode int, err error) {
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "OAuth "+c.oAuth)

	resp, err := c.client.Do(request)
	if err != nil {
		return resp.StatusCode, err
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	err = json.NewDecoder(resp.Body).Decode(&response)

	return statusCode, err
}

func (c *Client) GetDiskInfo(ctx context.Context) (disk *Disk, statusCode int, err error) {
	req, err := http.NewRequest("GET", c.baseURl, nil)
	if err != nil {
		return nil, statusCode, err
	}

	req = req.WithContext(ctx)

	diskResp := DiskResponse{}

	if statusCode, err = c.sendReq(req, &diskResp); err != nil {
		return nil, statusCode, err
	}

	if statusCode != 200 {
		err = fmt.Errorf(diskResp.Error.String())
	}

	return &diskResp.Disk, statusCode, err

}

func (c *Client) GetFiles(ctx context.Context) (resources *ResourceList, statusCode int, err error) {
	req, err := http.NewRequest("GET", c.baseURl+resourePath+"?path=app:/", nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)

	resourceResponse := ResourceResponse{}

	if statusCode, err = c.sendReq(req, &resourceResponse); err != nil {
		return
	}

	if statusCode != 200 {
		err = fmt.Errorf(resourceResponse.Error.String())
	}

	resources = &resourceResponse.ResourceList

	return
}

// func (c *Client) GetFiles(ctx context.Context, limit int) (*ResourceList, error) {
// 	req, err := http.NewRequest("GET", c.baseURl+resourePath+"?path=app:/", nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	q := req.URL.Query()
// 	q.Set("limit", strconv.Itoa(limit))
// 	req.URL.RawQuery = q.Encode()

// 	req = req.WithContext(ctx)

// 	filesResourceList := ResourceList{}

// 	if err = c.sendRequest(req, &filesResourceList); err != nil {
// 		return nil, err
// 	}

// 	return &filesResourceList, nil
// }

func (c *Client) MakeFolder(path string, ctx context.Context) (statusCode int, err error) {
	req, err := http.NewRequest("PUT", c.baseURl+resourePath+"?path="+path, nil)
	if err != nil {
		return statusCode, err
	}

	req = req.WithContext(ctx)

	resp := Response{}

	if statusCode, err = c.sendReq(req, &resp); err != nil {
		return statusCode, err
	}

	if statusCode != 201 {
		err = fmt.Errorf(resp.Error.String())
	}

	return statusCode, err
}

func (c *Client) Copy(from, path string, ctx context.Context) (statusCode int, err error) {
	req, err := http.NewRequest("POST", c.baseURl+resourePath+"copy?from="+from+"&path="+path, nil)
	if err != nil {
		return statusCode, err
	}

	req = req.WithContext(ctx)

	resp := Response{}

	if statusCode, err = c.sendReq(req, &resp); err != nil {
		return statusCode, err
	}

	if statusCode != 201 && statusCode != 202 {
		err = fmt.Errorf(resp.Error.String())
	}

	return statusCode, err
}

func (c *Client) GetDownloadLink(path string, ctx context.Context) (href string, statusCode int, err error) {
	req, err := http.NewRequest("GET", c.baseURl+resourePath+"download?path="+path, nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)

	resp := Response{}

	if statusCode, err = c.sendReq(req, &resp); err != nil {
		return
	}

	if statusCode != 200 {
		err = fmt.Errorf(resp.Error.String())
	}

	href = resp.Success.Href

	return
}

func (c *Client) GetUploadLink(path string, overwrite bool, ctx context.Context) (href string, statusCode int, err error) {
	req, err := http.NewRequest("GET", c.baseURl+resourePath+"upload?path="+path+"&overwrite="+strconv.FormatBool(overwrite), nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)

	resp := Response{}

	if statusCode, err = c.sendReq(req, &resp); err != nil {
		return
	}

	if statusCode != 200 {
		err = fmt.Errorf(resp.Error.String())
	}

	href = resp.Success.Href

	return
}

func (c *Client) UploadFile(path string, overwrite bool, ctx context.Context) (href string, statusCode int, err error) {
	// req, err := http.NewRequest("POST", c.baseURl+resourePath+"upload?path="+path+"&overwrite="+strconv.FormatBool(overwrite), nil)
	// if err != nil {
	// 	return
	// }
	return
}

func (c *Client) GetOperation(operation_id string, ctx context.Context) (status string, statusCode int, err error) {
	req, err := http.NewRequest("GET", c.baseURl+"operations/"+operation_id, nil)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)

	resp := StatusResponse{}

	if statusCode, err = c.sendReq(req, &resp); err != nil {
		return
	}

	if statusCode != 200 {
		err = fmt.Errorf(resp.Error.String())
	}

	status = resp.Status

	return
}
