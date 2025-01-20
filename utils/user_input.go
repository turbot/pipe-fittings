package utils

import (
	"fmt"
	"log"
	"strings"
)

func UserConfirmationWithDefault(warning string, defaultValue bool) bool {
	defString := "n"
	if defaultValue {
		defString = "y"
	}
	fmt.Println(fmt.Sprintf("%s (%s)", warning, defString)) //nolint:forbidigo // Console output

	var userConfirm string
	_, err := fmt.Scanf("%s", &userConfirm)
	if err != nil {
		if err.Error() == "unexpected newline" {
			return defaultValue
		}
		log.Fatal(err)
	}

	return strings.ToUpper(userConfirm) == "Y"
}

// UserConfirmation displays the warning message and asks the user for input
// regarding whether to continue or not
func UserConfirmation(warning string) bool {
	fmt.Println(warning) //nolint:forbidigo // Console output
	var userConfirm string
	_, err := fmt.Scanf("%s", &userConfirm)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToUpper(userConfirm) == "Y"
}
