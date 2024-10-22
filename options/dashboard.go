package options

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
)

// Dashboard options are deprecated and removed. They are kept here for compatibility with old configurations.
type Dashboard struct {
	// workspace profile
	Browser *bool `hcl:"browser" cty:"profile_dashboard_browser"`
}

func (d *Dashboard) SetBaseProperties(otherOptions Options) {
	if helpers.IsNil(otherOptions) {
		return
	}
	if o, ok := otherOptions.(*Dashboard); ok {
		if d.Browser == nil && o.Browser != nil {
			d.Browser = o.Browser
		}
	}
}

// ConfigMap creates a config map that can be merged with viper
func (d *Dashboard) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if d.Browser != nil {
		res[constants.ArgBrowser] = d.Browser
	}
	return res
}

// Merge :: merge other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (d *Dashboard) Merge(otherOptions Options) {
	if _, ok := otherOptions.(*Dashboard); !ok {
		return
	}
	switch o := otherOptions.(type) {
	case *Dashboard:
		if o.Browser != nil {
			d.Browser = o.Browser
		}
	}
}

func (d *Dashboard) String() string {
	if d == nil {
		return ""
	}
	var str []string
	if d.Browser == nil {
		str = append(str, "  Browser: nil")
	} else {
		str = append(str, fmt.Sprintf("  Browser: %v", *d.Browser))
	}
	return strings.Join(str, "\n")
}
