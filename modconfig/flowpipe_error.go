package modconfig

import (
	"math"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const (
	DefaultMaxAttempts = 3
	DefaultStrategy    = "constant"
	DefaultMinInterval = 1000
	DefaultMaxInterval = 10000
)

type RetryConfig struct {
	// circular link to its "parent"
	PipelineStepBase *PipelineStepBase `json:"-"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`

	If          *bool   `json:"if"`
	MaxAttempts *int64  `json:"max_attempts,omitempty" hcl:"max_attempts,optional" cty:"max_attempts"`
	Strategy    *string `json:"strategy,omitempty" hcl:"strategy,optional" cty:"strategy"`
	MinInterval *int64  `json:"min_interval,omitempty" hcl:"min_interval,optional" cty:"min_interval"`
	MaxInterval *int64  `json:"max_interval,omitempty" hcl:"max_interval,optional" cty:"max_interval"`
}

func NewRetryConfig(p *PipelineStepBase) *RetryConfig {
	return &RetryConfig{
		PipelineStepBase:     p,
		UnresolvedAttributes: make(map[string]hcl.Expression),
	}
}

func (r *RetryConfig) Equals(other *ErrorConfig) bool {
	return true
}

func (r *RetryConfig) AppendDependsOn(dependsOn ...string) {
	r.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (r *RetryConfig) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (r *RetryConfig) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	r.UnresolvedAttributes[name] = expr
}

func (r *RetryConfig) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeIf:
			r.AddUnresolvedAttribute(name, attr.Expr)
		case schema.AttributeTypeMaxAttempts:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, r)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				valInt, stepDiags := hclhelpers.CtyToInt64(val)
				if len(stepDiags) > 0 {
					diags = append(diags, stepDiags...)
					continue
				}

				r.MaxAttempts = valInt
			}
		case schema.AttributeTypeStrategy:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, r)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				valStr, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid strategy",
						Detail:   "strategy must be a string",
						Subject:  &attr.Range,
					})
					continue
				}

				r.Strategy = &valStr
			}

		case schema.AttributeTypeMinInterval:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, r)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				valInt, stepDiags := hclhelpers.CtyToInt64(val)
				if len(stepDiags) > 0 {
					diags = append(diags, stepDiags...)
					continue
				}

				r.MinInterval = valInt
			}

		case schema.AttributeTypeMaxInterval:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, r)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				valInt, stepDiags := hclhelpers.CtyToInt64(val)
				if len(stepDiags) > 0 {
					diags = append(diags, stepDiags...)
					continue
				}

				r.MaxInterval = valInt
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Unsupported attribute '" + name + "' in retry block",
				Subject:  &attr.Range,
			})
		}
	}

	moreDiags := r.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	return diags
}

func (r *RetryConfig) ResolveSettings() (int, string, int, int) {
	maxAttempts := r.MaxAttempts
	maxInterval := r.MaxInterval
	minInterval := r.MinInterval
	strategy := r.Strategy

	if maxAttempts == nil {
		maxAttempts = utils.ToPointer(int64(DefaultMaxAttempts))
	}
	if maxInterval == nil {
		maxInterval = utils.ToPointer(int64(DefaultMaxInterval))
	}

	if minInterval == nil {
		minInterval = utils.ToPointer(int64(DefaultMinInterval))
	}

	if strategy == nil {
		strategy = utils.ToPointer(DefaultStrategy)
	}

	return int(*maxAttempts), *strategy, int(*minInterval), int(*maxInterval)

}

// The first attempt is the first time the operation is tried, NOT the first
// retry.
//
// The first retry is the 2nd attempt
func (r *RetryConfig) CalculateBackoff(attempt int) time.Duration {

	if attempt <= 1 {
		return time.Duration(0)
	}

	_, strategy, minInterval, maxInterval := r.ResolveSettings()

	maxDuration := time.Duration(maxInterval) * time.Millisecond

	if strategy == "linear" {
		duration := time.Duration(minInterval*(attempt-1)) * time.Millisecond
		return min(duration, maxDuration)
	}

	if strategy == "exponential" {
		if attempt == 2 {
			return time.Duration(minInterval) * time.Millisecond
		}

		// The multiplier factor, usually 2 for exponential growth.
		factor := 2

		// Calculate the delay as baseInterval * 2^(attempt-1).
		// We subtract 1 from attempt to make the first attempt have no delay if desired.
		delay := float64(minInterval) * math.Pow(float64(factor), float64(attempt-2))

		duration := time.Duration(delay) * time.Millisecond
		if duration < 0 {
			return maxDuration
		}

		return min(duration, maxDuration)
	}

	return time.Duration(minInterval) * time.Millisecond
}

func (r *RetryConfig) Validate() hcl.Diagnostics {

	maxAttempts, strategy, minInterval, maxInterval := r.ResolveSettings()

	diags := hcl.Diagnostics{}
	if strategy != "constant" && strategy != "exponential" && strategy != "linear" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid retry strategy",
			Detail:   "Valid values are constant, exponential or linear",
		})
	}

	if maxAttempts > 3*100 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid max_attempts",
			Detail:   "max_attempts must be less than 300",
		})
	}

	if minInterval > 1000*100 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid min_interval",
			Detail:   "min_interval must be less than 100000",
		})
	}

	if minInterval < 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid min_interval",
			Detail:   "min_interval must be greater than 0",
		})
	}

	if maxInterval > 10000*100 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid max_interval",
			Detail:   "max_interval must be less than 1000000",
		})
	}

	if maxInterval < 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid max_interval",
			Detail:   "max_interval must be greater than 0",
		})
	}

	if minInterval >= maxInterval {
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
