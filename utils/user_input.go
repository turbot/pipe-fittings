package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func UserConfirmationWithDefault(message string, defaultValue bool) bool {
	defString := "n"
	if defaultValue {
		defString = "y"
	}
	fmt.Printf("%s (%s): ", message, defString) //nolint: forbidigo // No newline here for cleaner input prompt

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err) //nolint: forbidigo // Console output
		return defaultValue
	}

	// Trim whitespace and convert input to uppercase
	input = strings.TrimSpace(strings.ToUpper(input))

	// Default behavior if no input is provided
	if input == "" {
		return defaultValue
	}

	return input == "Y"
}

// UserConfirmation displays the warning message and asks the user for input
// regarding whether to continue or not
func UserConfirmation(message string) bool {
	fmt.Println(message) //nolint:forbidigo // Console output

	// Use bufio to read the input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err) //nolint:forbidigo // Console output
		return false
	}

	// Trim whitespace and normalize case
	input = strings.TrimSpace(strings.ToUpper(input))

	// Return true if the input is "Y", otherwise false
	return input == "Y"
}
