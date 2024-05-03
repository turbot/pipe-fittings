package error_helpers

import (
	"errors"
	"fmt"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/perr"
)

func MissingCloudTokenError() error {
	return fmt.Errorf("Not authenticated for Turbot Pipes.\nPlease run %s or setup a token.", constants.Bold(fmt.Sprintf("%s login", app_specific.AppName)))
}
func InvalidCloudTokenError() error {
	return fmt.Errorf("Invalid token.\nPlease run %s or setup a token.", constants.Bold(constants.Bold(fmt.Sprintf("%s login", app_specific.AppName))))
}

var InvalidStateError = errors.New("invalid state")

func MergeErrors(errs []error) []string {
	var errStrs []string
	for _, err := range errs {
		errModel, ok := err.(perr.ErrorModel)
		if ok {
			errStrs = append(errStrs, errModel.Detail)
		} else {
			errStrs = append(errStrs, err.Error())
		}
	}

	return errStrs
}
