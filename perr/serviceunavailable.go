package perr

import (
	"errors"
	"net/http"
)

const (
	ErrorCodeServiceUnavailable = "error_service_unavailable"
)

func ServiceUnavailable() ErrorModel {
	return ServiceUnavailableWithMessage("Service Unavailable")
}

func ServiceUnavailableWithMessage(msg string) ErrorModel {
	return ErrorModel{
		Instance: reference(),
		Type:     ErrorCodeServiceUnavailable,
		Title:    "Service Unavailable",
		Detail:   msg,
		Status:   http.StatusServiceUnavailable,
	}
}

func IsServiceUnavailable(err error) bool {
	var e ErrorModel
	ok := errors.As(err, &e)
	return ok && e.Status == http.StatusServiceUnavailable
}
