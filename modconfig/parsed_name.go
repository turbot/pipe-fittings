package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type ParsedResourceName struct {
	Mod      string
	ItemType string
	Name     string
}

func ParseResourceName(fullName string) (res *ParsedResourceName, err error) {
	if fullName == "" {
		return &ParsedResourceName{}, nil
	}
	res = &ParsedResourceName{}

	parts := strings.Split(fullName, ".")

	switch len(parts) {
	case 0:
		err = perr.BadRequestWithMessage("empty name passed to ParseResourceName")
	case 1:
		res.Name = parts[0]
	case 2:
		res.ItemType = parts[0]
		res.Name = parts[1]
	case 3:
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
	case 4:
		// this only applies for Triggers and Integration (as of 2023/09/13)
		// mod_name.trigger.schedule.trigger__name
		// mod_name.integration.slack.integration__name
		if parts[1] != schema.BlockTypeTrigger && parts[1] != schema.BlockTypeIntegration && parts[1] != schema.BlockTypeCredential {
			err = perr.BadRequestWithMessage(fmt.Sprintf("invalid name passed to ParseResourceName '%s' ", fullName))
		}
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2] + "." + parts[3]
	default:
		err = perr.BadRequestWithMessage(fmt.Sprintf("invalid name passed to ParseResourceName '%s'", fullName))
	}
	if !schema.IsValidResourceItemType(res.ItemType) {
		err = perr.BadRequestWithMessage("not a valid resource type passed to ParseResourceName '" + fullName + "' (" + res.ItemType + ")")
	}
	return
}

func (p *ParsedResourceName) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedResourceName) ToFullName() string {
	return BuildFullResourceName(p.Mod, p.ItemType, p.Name)
}
func (p *ParsedResourceName) ToFullNameWithMod(mod string) string {
	if p.Mod != "" {
		return p.ToFullName()
	}
	return BuildFullResourceName(mod, p.ItemType, p.Name)
}

func BuildFullResourceName(mod, blockType, name string) string {
	return fmt.Sprintf("%s.%s.%s", mod, blockType, name)
}

// UnqualifiedResourceName removes the mod prefix from the given name
func UnqualifiedResourceName(fullName string) string {
	parts := strings.Split(fullName, ".")
	switch len(parts) {
	case 3:
		return strings.Join(parts[1:], ".")
	default:
		return fullName
	}
}

func BuildModResourceName(blockType, name string) string {
	return fmt.Sprintf("%s.%s", blockType, name)
}
