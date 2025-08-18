package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/schema"
	"github.com/turbot/pipe-helpers/perr"
)

// ResourceNameParseFunc provides a mechanism for an app using pipe-fittings to override the default resource name parsing behavior.
// the default is to parse the resource name as <mod>.<block_type>.<name>.
// Tailpipe (for example) does not have mods, and some resources have subtypes. So, it provides its own resource name parser
// but for backwards compatibility, we provide a way to plug a new parser into the existing struct.
var ResourceNameParseFunc = parseResourceNameWithMod

type ResourceNameParser interface {
	ToResourceName() string
	ToFullName() string
	ToFullNameWithMod(mod string) string
	GetMod() string
	GetItemType() string
	GetName() string
	GetSubType() string
}

// ParsedResourceName is a container struct which holds the parsed resource name.
type ParsedResourceName struct {
	impl     ResourceNameParser
	Mod      string
	ItemType string
	Name     string
}

func ParseResourceName(fullName string) (*ParsedResourceName, error) {

	parsed, err := ResourceNameParseFunc(fullName)
	if err != nil {
		return nil, err
	}
	res := &ParsedResourceName{
		impl:     parsed,
		Mod:      parsed.GetMod(),
		ItemType: parsed.GetItemType(),
		Name:     parsed.GetName(),
	}
	return res, nil
}

func (p ParsedResourceName) ToResourceName() string {
	return p.impl.ToResourceName()
}

func (p ParsedResourceName) ToFullName() string {
	return p.impl.ToFullName()
}

func (p ParsedResourceName) ToFullNameWithMod(mod string) string {
	return p.impl.ToFullNameWithMod(mod)
}

func (p ParsedResourceName) GetSubType() string {
	return p.impl.GetSubType()
}

func parseResourceNameWithMod(fullName string) (ResourceNameParser, error) {
	p := &ParsedResourceNameWithMod{}
	if fullName == "" {
		return p, nil
	}
	var err error

	parts := strings.Split(fullName, ".")

	switch len(parts) {
	case 0:
		err = perr.BadRequestWithMessage("empty name passed to ParseResourceName")
	case 1:
		p.Name = parts[0]
	case 2:
		p.ItemType = parts[0]
		p.Name = parts[1]
	case 3:
		p.Mod = parts[0]
		p.ItemType = parts[1]
		p.Name = parts[2]
	case 4:
		// this only applies for Triggers and Integration (as of 2023/09/13)
		// mod_name.trigger.schedule.trigger__name
		// mod_name.integration.slack.integration__name
		if parts[1] != schema.BlockTypeTrigger && parts[1] != schema.BlockTypeIntegration && parts[1] != schema.BlockTypeCredential {
			err = perr.BadRequestWithMessage(fmt.Sprintf("invalid name passed to ParseResourceName '%s' ", fullName))
		}
		p.Mod = parts[0]
		p.ItemType = parts[1]
		p.Name = parts[2] + "." + parts[3]
	default:
		err = perr.BadRequestWithMessage(fmt.Sprintf("invalid name passed to ParseResourceName '%s'", fullName))
	}

	return p, err
}

// ParsedResourceNameWithMod is the default resource name parser implementation
// which handles resource names with a mod prefix.
type ParsedResourceNameWithMod struct {
	Mod      string
	ItemType string
	Name     string
}

func (p *ParsedResourceNameWithMod) GetMod() string {
	return p.Mod
}

func (p *ParsedResourceNameWithMod) GetItemType() string {
	return p.ItemType
}

func (p *ParsedResourceNameWithMod) GetSubType() string {
	return ""
}

func (p *ParsedResourceNameWithMod) GetName() string {
	return p.Name
}

func (p *ParsedResourceNameWithMod) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedResourceNameWithMod) ToFullName() string {
	if p.Mod == "" {
		return p.ToResourceName()
	}
	return buildFullResourceName(p.Mod, p.ItemType, p.Name)
}

func (p *ParsedResourceNameWithMod) ToFullNameWithMod(mod string) string {
	// use existing mod if set
	if p.Mod != "" {
		return p.ToFullName()
	}
	return buildFullResourceName(mod, p.ItemType, p.Name)
}

func buildFullResourceName(mod, blockType, name string) string {
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
