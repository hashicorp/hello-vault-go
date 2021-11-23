package util

import (
	"net/http"
)

var NotFoundError = Error{
	Code:    http.StatusNotFound,
	Message: "Resource not found.",
}

var InternalServerError = Error{
	Code:    http.StatusInternalServerError,
	Message: "Our technical team has been notified.",
}

type Error struct {
	Internal error
	Code     int
	Message  string
	Response interface{}
}

func (e Error) Error() string {
	return e.Internal.Error()
}
