package go_request

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"io"
	"net/http"
	"testing"
)

func TestRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type test struct {
		name       string
		setup      func(*testing.T) []Option
		assertions func(*testing.T, []Option)
	}

	type target struct {
		SomeValue string
	}

	targetString := "Propane and propane accessories"
	returnedCode := 69
	tests := []test{
		{
			name: "happy path all options and target",
			setup: func(t *testing.T) []Option {
				targ := target{
					SomeValue: targetString,
				}
				b, err := json.Marshal(&targ)
				if err != nil {
					t.Fatal("failed to marshal target in setup")
				}
				return []Option{
					WithScheme(HTTPS),
					WithMethod(Get),
					WithBody(b),
					WithHost("strickland-propane.com"),
					WithPath("dangit", "bobby"),
					WithClient(&mockedClient{}),
					WithQueryArgs(map[string][]string{
						"query": {"arg1", "arg2"},
					}),
					WithHeaders(map[string][]string{
						"traceability": {"path1", "path2"},
					}),
					WithResponseHandler(func(response *http.Response) error {
						response.StatusCode = returnedCode
						response.Body = io.NopCloser(bytes.NewBuffer(b))
						return nil
					}),
				}
			},
			assertions: func(t *testing.T, opts []Option) {
				var targ target
				ctx := context.Background()
				b, err := NewRequester(opts...)
				if err != nil {
					t.Fatal("err was unexpectedly not nil when building a request")
				}
				resp, err := b.Make(ctx, &targ)
				if err != nil {
					t.Fatal("err was unexpectedly not nil when making a request")
				}
				if resp.StatusCode != returnedCode {
					t.Errorf("wrong status code, got %d, want %d", resp.StatusCode, returnedCode)
				}
				if targ.SomeValue != targetString {
					t.Errorf("wrong target string, got %s, want %s", targ.SomeValue, targetString)
				}
			},
		},
		{
			name: "happy path defaults",
			setup: func(t *testing.T) []Option {
				return []Option{
					WithHost("strickland-propane.com"),
					WithClient(&mockedClient{status: 200}),
				}
			},
			assertions: func(t *testing.T, opts []Option) {
				ctx := context.Background()
				b, err := NewRequester(opts...)
				if err != nil {
					t.Fatal("err was unexpectedly not nil when building a request")
				}
				resp, err := b.Make(ctx, nil)
				if err != nil {
					t.Fatal("err was unexpectedly not nil when making a request")
				}
				if resp.StatusCode != http.StatusOK {
					t.Errorf("wrong status code, got %d, want %d", resp.StatusCode, returnedCode)
				}
			},
		},
		{
			name: "missing host will cause error",
			setup: func(t *testing.T) []Option {
				return []Option{
					WithClient(&mockedClient{status: 200}),
				}
			},
			assertions: func(t *testing.T, opts []Option) {
				_, err := NewRequester(opts...)
				if err == nil {
					t.Fatal("an error was expected")
				}
			},
		},
		{
			name: "sad path default handler with unacceptable status code",
			setup: func(t *testing.T) []Option {
				return []Option{
					WithHost("strickland-propane.com"),
					WithClient(&mockedClient{status: 420}),
				}
			},
			assertions: func(t *testing.T, opts []Option) {
				ctx := context.Background()
				b, err := NewRequester(opts...)
				if err != nil {
					t.Fatal("err was unexpectedly not nil when building a request")
				}
				_, err = b.Make(ctx, nil)
				if err == nil {
					t.Fatal("an error was expected due to bad status code")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.setup(t)
			tt.assertions(t, opts)
		})
	}
}

// mock helpers
type mockedClient struct {
	target any
	status int
	error
}

func (m *mockedClient) Do(_ *http.Request) (*http.Response, error) {
	r := &http.Response{
		StatusCode: m.status,
	}

	out, err := json.Marshal(m.target)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(out))

	return r, m.error
}
