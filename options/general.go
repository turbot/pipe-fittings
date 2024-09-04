package options

import (
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"strings"

	"github.com/turbot/pipe-fittings/constants"
)

// TODO #steampipe kai check max parallel - do we need separate general options
type General struct {
	UpdateCheck *string `hcl:"update_check" cty:"update_check"`
	MaxParallel *int    `hcl:"max_parallel" cty:"max_parallel"`
	Telemetry   *string `hcl:"telemetry" cty:"telemetry"`
	LogLevel    *string `hcl:"log_level" cty:"log_level"`
	MemoryMaxMb *int    `hcl:"memory_max_mb" cty:"memory_max_mb"`
}

// TODO KAI what is the difference between merge and SetBaseProperties
func (s *General) SetBaseProperties(otherOptions Options) {
	if helpers.IsNil(otherOptions) {
		return
	}
	if o, ok := otherOptions.(*General); ok {
		if s.UpdateCheck == nil && o.UpdateCheck != nil {
			s.UpdateCheck = o.UpdateCheck
		}
		if s.MaxParallel == nil && o.MaxParallel != nil {
			s.MaxParallel = o.MaxParallel
		}
		if s.Telemetry == nil && o.Telemetry != nil {
			s.Telemetry = o.Telemetry
		}
		if s.LogLevel == nil && o.LogLevel != nil {
			s.LogLevel = o.LogLevel
		}
		if s.MemoryMaxMb == nil && o.MemoryMaxMb != nil {
			s.MemoryMaxMb = o.MemoryMaxMb
		}

	}
}

// ConfigMap creates a config map that can be merged with viper
func (g *General) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if g.UpdateCheck != nil {
		res[constants.ArgUpdateCheck] = g.UpdateCheck
	}
	if g.Telemetry != nil {
		res[constants.ArgTelemetry] = g.Telemetry
	}
	if g.MaxParallel != nil {
		res[constants.ArgMaxParallel] = g.MaxParallel
	}
	if g.LogLevel != nil {
		res[constants.ArgLogLevel] = g.LogLevel
	}
	if g.MemoryMaxMb != nil {
		res[constants.ArgMemoryMaxMb] = g.MemoryMaxMb
	}

	return res
}

// Merge merges other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (g *General) Merge(otherOptions Options) {
	// TODO KAI this seems incomplete - check all merge
	// also who uses this???
	switch o := otherOptions.(type) {
	case *General:
		if o.UpdateCheck != nil {
			g.UpdateCheck = o.UpdateCheck
		}
	}
}

func (g *General) String() string {
	if g == nil {
		return ""
	}
	var str []string
	if g.UpdateCheck == nil {
		str = append(str, "  UpdateCheck: nil")
	} else {
		str = append(str, fmt.Sprintf("  UpdateCheck: %s", *g.UpdateCheck))
	}

	if g.MaxParallel == nil {
		str = append(str, "  MaxParallel: nil")
	} else {
		str = append(str, fmt.Sprintf("  MaxParallel: %d", *g.MaxParallel))
	}

	if g.Telemetry == nil {
		str = append(str, "  Telemetry: nil")
	} else {
		str = append(str, fmt.Sprintf("  Telemetry: %s", *g.Telemetry))
	}
	if g.LogLevel == nil {
		str = append(str, "  LogLevel: nil")
	} else {
		str = append(str, fmt.Sprintf("  LogLevel: %s", *g.LogLevel))
	}

	if g.MemoryMaxMb == nil {
		str = append(str, "  MemoryMaxMb: nil")
	} else {
		str = append(str, fmt.Sprintf("  MemoryMaxMb: %d", *g.MemoryMaxMb))
	}
	return strings.Join(str, "\n")
}
