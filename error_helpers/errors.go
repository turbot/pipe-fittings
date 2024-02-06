package error_helpers

import (
	"errors"
	"fmt"

	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
)

var MissingCloudTokenError = fmt.Errorf("Not authenticated for Turbot Pipes.\nPlease run %s or setup a token.", constants.Bold(fmt.Sprintf("%s login", app_specific.AppName)))
var InvalidCloudTokenError = fmt.Errorf("Invalid token.\nPlease run %s or setup a token.", constants.Bold(constants.Bold(fmt.Sprintf("%s login", app_specific.AppName))))
var InvalidStateError = errors.New("invalid state")

func MergeErrors(errs []error) []string {
	var errStrs []string
	for _, err := range errs {
		errStrs = append(errStrs, err.Error())
	}

	return errStrs
}
