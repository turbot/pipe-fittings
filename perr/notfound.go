package perr

import (
	"fmt"
	"net/http"
)

const (
	ErrorCodeNotFound        = "error_not_found"
	ErrorCodeTriggerDisabled = "error_trigger_disabled"
)

func NotFoundWithMessage(msg string) ErrorModel {
	return NotFoundWithMessageAndType(ErrorCodeNotFound, msg)
}

func NotFoundWithMessageAndType(errorType string, msg string) ErrorModel {
	id := reference()
	e := ErrorModel{
		Instance: id,
		Type:     errorType,
		Title:    "Not Found",
		Status:   http.StatusNotFound,
		Detail:   msg,
	}
	return e
}

func NotFound(itemType string, id string) ErrorModel {
	e := ErrorModel{
		Instance: reference(),
		Type:     ErrorCodeNotFound,
		Title:    "Not Found",
		Status:   http.StatusNotFound,
	}
	if id != "" {
		e.Detail = fmt.Sprintf("%s = %s.", itemType, id)
	}
	return e
}

func IsNotFound(err error) bool {
	e, ok := err.(ErrorModel)
	return ok && (e.Status == http.StatusNotFound)
}
