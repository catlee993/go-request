package go_request

import (
	"context"
	"net/http"
)

type Client interface {
	Do(r *http.Request) (*http.Response, error)
}

type Requester[T any] interface {
	Make(ctx context.Context, responseTarget T) (*http.Response, error)
}
