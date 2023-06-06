package hr

import (
	"fmt"
	"io"
	"net/http"
)

type Error struct {
	Code   int
	Detail string
}

func (e Error) Error() string {
	return fmt.Sprintf(`{"code":%d,"detail":"%s"}`, e.Code, e.Detail)
}

func (e Error) WriteTo(w io.Writer) (int64, error) {
	if rw, ok := w.(http.ResponseWriter); ok {
		code := e.Code
		if code >= 1000 {
			// when it's self-defined error code, we use a 200 status code
			// in the response header instead, leaving the error handling
			// to the client.
			code = http.StatusOK
		}
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(code)
	}
	n, err := fmt.Fprint(w, e.Error())
	return int64(n), err
}

func ParseError(err error) Error {
	if herr, ok := err.(Error); ok {
		return herr
	}
	return Error{
		Code:   http.StatusInternalServerError,
		Detail: err.Error(),
	}
}

func BadRequest(format string, v ...interface{}) Error {
	return Error{Code: http.StatusBadRequest, Detail: fmt.Sprintf(format, v...)}
}

func Forbidden(format string, v ...interface{}) Error {
	return Error{Code: http.StatusForbidden, Detail: fmt.Sprintf(format, v...)}
}

func Unauthorized(format string, v ...interface{}) Error {
	return Error{Code: http.StatusUnauthorized, Detail: fmt.Sprintf(format, v...)}
}

func NotFound(format string, v ...interface{}) Error {
	return Error{Code: http.StatusNotFound, Detail: fmt.Sprintf(format, v...)}
}

func InternalServerError(format string, v ...interface{}) Error {
	return Error{Code: http.StatusInternalServerError, Detail: fmt.Sprintf(format, v...)}
}
