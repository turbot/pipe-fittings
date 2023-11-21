package perr

import (
	"fmt"
)

const (
	ErrorCodeExecutionError = "error_execution"
	StatusExecutionError    = 460
)

func ExecutionError(itemType string, id string) ErrorModel {
	e := ErrorModel{
		Instance: reference(),
		Type:     ErrorCodeExecutionError,
		Title:    "Execution Error",
		Status:   StatusExecutionError,
	}
	if id != "" {
		e.Detail = fmt.Sprintf("%s = %s.", itemType, id)
	}
	return e
}

func ExecutionErrorWithMessage(msg string) ErrorModel {
	return ExecutionErrorWithTypeAndMessage(ErrorCodeExecutionError, msg)
}

func ExecutionErrorWithTypeAndMessage(errorType string, msg string) ErrorModel {
	e := ErrorModel{
		Instance: reference(),
		Type:     errorType,
		Title:    "Execution Error",
		Status:   StatusExecutionError,
		Detail:   msg,
	}
	return e
}

func IsExecutionError(err error) bool {
	e, ok := err.(ErrorModel)
	return ok && e.Status == StatusExecutionError
}
