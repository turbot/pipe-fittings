package modconfig

import "github.com/hashicorp/hcl/v2"

type RetryConfig struct {
	If      *bool `json:"if,omitempty" hcl:"if,optional" cty:"if"`
	Retries int   `json:"retries" hcl:"retries" cty:"retries"`
}

type ThrowConfig struct {
	If             bool     `json:"if" hcl:"if" cty:"if"`
	Message        *string  `json:"message,omitempty" hcl:"message,optional" cty:"message"`
	Unresolved     bool     `json:"unresolved"`
	UnresolvedBody hcl.Body `json:"-"`
}
