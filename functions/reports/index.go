package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type APIGatewayRequest struct {
	OperationID string `json:"operationId"`
	Resource    string `json:"resource"`

	HTTPMethod string `json:"httpMethod"`

	Path           string            `json:"path"`
	PathParameters map[string]string `json:"pathParameters"`

	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`

	QueryStringParameters           map[string]string   `json:"queryStringParameters"`
	MultiValueQueryStringParameters map[string][]string `json:"multiValueQueryStringParameters"`

	Parameters           map[string]string   `json:"parameters"`
	MultiValueParameters map[string][]string `json:"multiValueParameters"`

	Body            string `json:"body"`
	IsBase64Encoded bool   `json:"isBase64Encoded,omitempty"`

	RequestContext interface{} `json:"requestContext"`
}

type APIGatewayResponse struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
	IsBase64Encoded   bool                `json:"isBase64Encoded,omitempty"`
}

type Request struct {
	Name     string `json:"name"`
	UserId   string `json:"user_id"`
	DateFrom string `json:"datefrom"`
	DateTo   string `json:"dateto"`
}

func Handler(ctx context.Context, event *APIGatewayRequest) (resp *APIGatewayResponse, err error) {
	req := &Request{}
	if err = json.Unmarshal([]byte(event.Body), &req); err != nil {
		return nil, fmt.Errorf("an error has occurred when parsing body: %v", err)
	}

	if db == nil {
		fmt.Println("connectDB: new connection")
		db, err = connectDB(context.Background(), os.Getenv("YDB_DSN"), "")
		if err != nil {
			return &APIGatewayResponse{
				StatusCode: 500,
			}, err
		}
	} else {
		fmt.Println("connectDB: already connected")
	}

	p := make(map[string]string)
	p["user_id"] = req.UserId
	p["datefrom"] = req.DateFrom
	p["dateto"] = req.DateTo

	res, err := db.GetStatementCatTotals(context.Background(), p)
	if err != nil {
		return &APIGatewayResponse{
			StatusCode: 500,
		}, err
	}
	resVal, err := json.Marshal(res)
	if err != nil {
		return &APIGatewayResponse{
			StatusCode: 500,
		}, err
	}

	return &APIGatewayResponse{
		StatusCode: 200,
		Body:       string(resVal),
	}, nil
}
