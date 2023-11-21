package modconfig

import (
	"time"

	"github.com/hashicorp/hcl/v2"
)

type RetryConfig struct {
	If          *bool  `json:"if,omitempty" hcl:"if,optional" cty:"if"`
	MaxAttempts int    `json:"max_attempts,omitempty" hcl:"max_attempts,optional" cty:"max_attempts"`
	Strategy    string `json:"strategy,omitempty" hcl:"strategy,optional" cty:"strategy"`
	MinInterval int    `json:"min_interval,omitempty" hcl:"min_interval,optional" cty:"min_interval"`
	MaxInterval int    `json:"max_interval,omitempty" hcl:"max_interval,optional" cty:"max_interval"`
}

func NewRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3, // TODO: should we have max attempts?
		Strategy:    "constant",
		MinInterval: 1000,
		MaxInterval: 10000,
	}
}

func (r *RetryConfig) CalculateBackoff(attempt int) time.Duration {

	maxDuration := time.Duration(r.MaxInterval) * time.Millisecond

	if r.Strategy == "linear" {
		duration := time.Duration(r.MinInterval*attempt) * time.Millisecond
		return min(duration, maxDuration)
	}

	if r.Strategy == "exponential" {
		duration := time.Duration(r.MinInterval*attempt*attempt) * time.Millisecond
		return min(duration, maxDuration)
	}

	return time.Duration(r.MinInterval) * time.Millisecond
}

func (r *RetryConfig) Validate() hcl.Diagnostics {

	diags := hcl.Diagnostics{}
	if r.Strategy != "constant" && r.Strategy != "exponential" && r.Strategy != "linear" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid retry strategy",
			Detail:   "Valid values are constant, exponential or linear",
		})
	}

	if r.MaxAttempts > 3*100 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid max_attempts",
			Detail:   "max_attempts must be less than 300",
		})
	}

	if r.MinInterval > 1000*100 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid min_interval",
			Detail:   "min_interval must be less than 100000",
		})
	}

	if r.MinInterval < 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid min_interval",
			Detail:   "min_interval must be greater than 0",
		})
	}

	if r.MaxInterval > 10000*100 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid max_interval",
			Detail:   "max_interval must be less than 1000000",
		})
	}

	if r.MaxInterval < 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid max_interval",
			Detail:   "max_interval must be greater than 0",
		})
	}

	if r.MinInterval >= r.MaxInterval {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid min_interval",
			Detail:   "min_interval must be less than max_interval",
		})
	}

	return diags
}

type ThrowConfig struct {
	If             bool     `json:"if" hcl:"if" cty:"if"`
	Message        *string  `json:"message,omitempty" hcl:"message,optional" cty:"message"`
	Unresolved     bool     `json:"unresolved"`
	UnresolvedBody hcl.Body `json:"-"`
}
