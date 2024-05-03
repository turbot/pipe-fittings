package options

import (
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
)

// TODO KAI this is Flowpipe specific - should it be named as such
type Server struct {
	Port   *int    `hcl:"port" cty:"port"`
	Listen *string `hcl:"listen" cty:"port"`
}

// TODO KAI what is the difference between merge and SetBaseProperties
func (s *Server) SetBaseProperties(otherOptions Options) {
	if helpers.IsNil(otherOptions) {
		return
	}
	if o, ok := otherOptions.(*Server); ok {
		if s.Port == nil && o.Port != nil {
			s.Port = o.Port
		}
		if s.Listen == nil && o.Listen != nil {
			s.Listen = o.Listen
		}
	}
}

// ConfigMap creates a config map that can be merged with viper
func (s *Server) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if s.Port != nil {
		res[constants.ArgPort] = s.Port
	}
	if s.Listen != nil {
		res[constants.ArgListen] = s.Listen
	}

	return res
}

// Merge merges other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (s *Server) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *Server:
		if o.Port != nil {
			s.Port = o.Port
		}
		if o.Listen != nil {
			s.Listen = o.Listen
		}
	}
}

func (s *Server) String() string {
	if s == nil {
		return ""
	}
	var str []string
	if s.Listen == nil {
		str = append(str, "  Listen: nil")
	} else {
		str = append(str, fmt.Sprintf("  Listen: %s", *s.Listen))
	}

	if s.Port == nil {
		str = append(str, "  Port: nil")
	} else {
		str = append(str, fmt.Sprintf("  Port: %d", *s.Port))
	}

	return strings.Join(str, "\n")
}
