package main

import (
	"context"
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

type Response struct {
	IsAuthorized bool        `json:"isAuthorized"`
	Context      interface{} `json:"context"`
}

func Handler(ctx context.Context, event *APIGatewayRequest) (r *Response, err error) {
	r = &Response{
		IsAuthorized: false,
		Context:      nil,
	}
	if event.Headers["Authorization"] == "Basic secretToken" {
		r.IsAuthorized = true
	}

	return r, err
}
