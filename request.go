package go_request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

type void struct{}

type ResponseHandler func(response *http.Response) error

type builder struct {
	client             Client
	req                *http.Request
	scheme             Scheme
	method             Method
	host, path         string
	headers, queryArgs map[string][]string
	body               []byte
	responseHandler    ResponseHandler
}

type Option func(*builder)

type Method string

const (
	Get    Method = "GET"
	Delete Method = "DELETE"
	Patch  Method = "PATCH"
	Post   Method = "POST"
	Put    Method = "PUT"
)

type Scheme string

const (
	HTTP  Scheme = "http"
	HTTPS Scheme = "https"
)

func WithMethod(method Method) Option {
	return func(b *builder) {
		b.method = method
	}
}

func WithScheme(scheme Scheme) Option {
	return func(b *builder) {
		b.scheme = scheme
	}
}

func WithHost(host string) Option {
	return func(b *builder) {
		b.host = host
	}
}

func WithPath(chunks ...string) Option {
	return func(b *builder) {
		if len(chunks) > 0 {
			b.path = path.Join(chunks...)

			return
		}
		b.path = chunks[0]
	}
}

func WithHeaders(headers map[string][]string) Option {
	return func(b *builder) {
		b.headers = headers
	}
}

func WithQueryArgs(queryArgs map[string][]string) Option {
	return func(b *builder) {
		b.queryArgs = queryArgs
	}
}

func WithBody(body []byte) Option {
	return func(b *builder) {
		b.body = body
	}
}

func WithClient(c Client) Option {
	return func(b *builder) {
		b.client = c
	}
}

func WithResponseHandler(rh ResponseHandler) Option {
	return func(b *builder) {
		b.responseHandler = rh
	}
}

// Make makes the request and manages any returned result.
func (b *builder) Make(ctx context.Context, target any) (*http.Response, error) {
	b.req.WithContext(ctx)
	resp, err := b.client.Do(b.req)
	if err != nil {
		return nil, err
	}

	rErr := b.responseHandler(resp)
	if rErr != nil {
		return resp, rErr
	}

	// there is no target, so just return the response as-is
	if target == nil {
		return resp, nil
	}

	// there is a target, so populate body to the target's location and return (or error)
	defer closeBody(resp.Body)

	return resp, json.NewDecoder(resp.Body).Decode(&target)
}

// NewRequester populates a builder based on argued options or defaults, if available
func NewRequester(opts ...Option) (Requester[any], error) {
	var b builder
	for _, o := range opts {
		o(&b)
	}

	// Request can't have a host
	if len(b.host) == 0 {
		return nil, fmt.Errorf("missing host from request builder: %v", b)
	}

	// Use default scheme if not argued
	if len(b.scheme) == 0 {
		b.scheme = defaultScheme
	}

	// Use default GET method if not argued
	if len(b.method) == 0 {
		b.method = defaultMethod
	}

	// URL is joined with scheme, host and path (if argued)
	url := formatURL(b.scheme, b.host, b.path)

	// Create the request and apply query args & headers (if applicable)
	req, err := http.NewRequest(string(b.method), url.String(), bytes.NewBuffer(b.body))
	if err != nil {
		return nil, err
	}

	if b.queryArgs != nil {
		addQueryArgs(req, b.queryArgs)
	}

	if b.headers != nil {
		addHeaders(req, b.headers)
	}

	// A client is necessary to make this request, so just a default/bare-bones client
	// here if not argued. I strongly suggest you define your own client to specify
	// transport/timeout, etc
	if b.client == nil {
		b.client = http.DefaultClient
	}

	if b.responseHandler == nil {
		b.responseHandler = defaultResponseHandler
	}

	b.req = req

	return &b, nil
}
