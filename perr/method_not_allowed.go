package perr

import (
	"net/http"
)

const (
	ErrorCodeMethodNotAllowed = "not_allowed"
)

func MethodNotAllowed() ErrorModel {
	return MethodNotAllowedWithMessage("Method Not Allowed.")
}

func MethodNotAllowedWithMessage(msg string) ErrorModel {
	return ErrorModel{
		Instance: reference(),
		Type:     ErrorCodeMethodNotAllowed,
		Title:    "Method Not Allowed",
		Detail:   msg,
		Status:   http.StatusMethodNotAllowed,
	}
}

func IsMethodNotAllowed(err error) bool {
	e, ok := err.(ErrorModel)
	return ok && e.Status == http.StatusMethodNotAllowed
}
