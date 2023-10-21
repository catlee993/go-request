package go_request

import "fmt"

type ResponseError struct {
	message string
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("request failed: %s", e.message)
}

func NewResponseError(statusCode int, body []byte) error {
	return &ResponseError{
		message: fmt.Sprintf("response not OK, status: %d, body: %s", statusCode, body),
	}
}

func IsResponseErr(err error) bool {
	_, ok := err.(ResponseError)
	return ok
}
