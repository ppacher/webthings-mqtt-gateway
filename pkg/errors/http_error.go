package errors

import (
	"encoding/json"
	"errors"
)

// HTTPError wraps the built-in error interface and adds
// a status-code method that may be used to specify an HTTP status
// code that should be used for the error
type HTTPError interface {
	error
	json.Marshaler

	StatusCode() int
}

type httpError struct {
	inner error
	code  int
}

func (err *httpError) Error() string {
	return err.inner.Error()
}

func (err *httpError) StatusCode() int {
	if err.code == 0 {
		return 500
	}

	return err.code
}

func (err *httpError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code":  err.code,
		"error": err.inner.Error(),
	})
}

func WrapWithStatus(code int, err error) error {
	if err == nil {
		return nil
	}

	return &httpError{
		inner: err,
		code:  code,
	}
}

func MayWrap(code int, err error) error {
	if _, ok := err.(HTTPError); ok {
		return err
	}

	return WrapWithStatus(code, err)
}

func NewWithStatus(code int, msg string) error {
	return WrapWithStatus(code, errors.New(msg))
}
