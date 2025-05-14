package hclhelpers

import (
	"errors"
	"fmt"
	"strings"
)

type HclTag struct {
	Tag   string
	Block bool
	// was optional specified in the hcl tag
	// (NOTE - the field may be optional even if not specified, so this is a pointer)
	Optional *bool
	// was remain specified in the hcl tag
	Remain bool
}

// NewHclTag creates a new HclTag from a string and validates it.
func NewHclTag(tag string) (HclTag, error) {
	// Split the tag into its components (assuming comma-separated)
	tagParts := strings.Split(tag, ",")

	// Ensure that the tag has either 1 or 2 parts (valid HCL tags should have at most 1 comma)
	if len(tagParts) > 2 {
		return HclTag{}, errors.New("invalid HCL tag: too many parts, only one comma is allowed")
	}

	// Initialize the HclTag struct with the field name
	hclTag := HclTag{
		Tag: tagParts[0], // The first part is always the tag name (e.g., field name or block name)
	}

	// If there's no modifier (just the tag name), return it as valid
	if len(tagParts) == 1 {
		return hclTag, nil
	}

	// Check if the second part is a valid modifier (either 'block' or 'optional')
	modifier := strings.TrimSpace(tagParts[1])
	switch modifier {
	case "block":
		hclTag.Block = true
	case "optional":
		optional := true
		hclTag.Optional = &optional
	case "remain":
		hclTag.Remain = true
	default:
		return HclTag{}, fmt.Errorf("invalid HCL tag: unknown modifier '%s', must be 'block' or 'optional'", modifier)
	}

	// Return the populated HclTag
	return hclTag, nil
}
