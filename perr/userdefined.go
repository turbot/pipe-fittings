package perr

import (
	"fmt"
)

const (
	ErrorCodeUserDefined = "error_user_defined"
	StatusUserDefined    = 461
)

func UserDefined(itemType string, id string) ErrorModel {
	e := ErrorModel{
		Instance: reference(),
		Type:     ErrorCodeUserDefined,
		Title:    "User Defined",
		Status:   StatusUserDefined,
	}
	if id != "" {
		e.Detail = fmt.Sprintf("%s = %s.", itemType, id)
	}
	return e
}

func UserDefinedWithMessage(msg string) ErrorModel {
	return UserDefinedWithTypeAndMessage(ErrorCodeUserDefined, msg)
}

func UserDefinedWithTypeAndMessage(errorType string, msg string) ErrorModel {
	e := ErrorModel{
		Instance: reference(),
		Type:     errorType,
		Title:    "User Defined",
		Status:   StatusUserDefined,
		Detail:   msg,
	}
	return e
}

func IsUserDefined(err error) bool {
	e, ok := err.(ErrorModel)
	return ok && e.Status == StatusUserDefined
}
