package go_request

import (
	"io"
	"net/http"
	"net/url"
)

const (
	defaultScheme = HTTP
	defaultMethod = Get
)

func defaultResponseHandler(resp *http.Response) error {
	successCodes := map[int]void{
		http.StatusOK:                   {},
		http.StatusCreated:              {},
		http.StatusAccepted:             {},
		http.StatusNonAuthoritativeInfo: {},
		http.StatusNoContent:            {},
		http.StatusResetContent:         {},
		http.StatusPartialContent:       {},
		http.StatusMultiStatus:          {},
		http.StatusAlreadyReported:      {},
		http.StatusIMUsed:               {},
	}

	// This default handler will assume 300 codes are acceptable
	redirectCodes := map[int]void{
		http.StatusMultipleChoices:   {},
		http.StatusMovedPermanently:  {},
		http.StatusFound:             {},
		http.StatusSeeOther:          {},
		http.StatusNotModified:       {},
		http.StatusUseProxy:          {},
		http.StatusTemporaryRedirect: {},
		http.StatusPermanentRedirect: {},
	}

	if _, ok := successCodes[resp.StatusCode]; !ok {
		if _, isRedirect := redirectCodes[resp.StatusCode]; isRedirect {
			return nil
		}

		defer closeBody(resp.Body)
		b, _ := io.ReadAll(resp.Body)
		return NewResponseError(resp.StatusCode, b)
	}

	return nil
}

func formatURL(scheme Scheme, host, path string) url.URL {
	return url.URL{
		Scheme: string(scheme),
		Host:   host,
		Path:   path,
	}
}

func addQueryArgs(r *http.Request, args map[string][]string) {
	q := r.URL.Query()
	for k, values := range args {
		for _, v := range values {
			q.Add(k, v)
		}
	}
	r.URL.RawQuery = q.Encode()
}

func addHeaders(r *http.Request, headers map[string][]string) {
	for k, values := range headers {
		for _, v := range values {
			r.Header.Add(k, v)
		}
	}
}

func closeBody(body io.ReadCloser) {
	_ = body.Close()
}
