package modconfig

import (
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const (
	HttpMethodGet    = "get"
	HttpMethodPost   = "post"
	HttpMethodPut    = "put"
	HttpMethodDelete = "delete"
	HttpMethodPatch  = "patch"
)

var ValidHttpMethods = []string{
	HttpMethodGet,
	HttpMethodPost,
	HttpMethodPut,
	HttpMethodDelete,
	HttpMethodPatch,
}

type PipelineStepHttp struct {
	PipelineStepBase

	Url             *string                `json:"url" binding:"required"`
	Method          *string                `json:"method,omitempty"`
	CaCertPem       *string                `json:"ca_cert_pem,omitempty"`
	Insecure        *bool                  `json:"insecure,omitempty"`
	RequestBody     *string                `json:"request_body,omitempty"`
	RequestHeaders  map[string]interface{} `json:"request_headers,omitempty"`
	BasicAuthConfig *BasicAuthConfig       `json:"basic_auth_config,omitempty"`
}

func (p *PipelineStepHttp) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepHttp)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Compare Url field
	if reflect.DeepEqual(p.Url, other.Url) {
		return false
	}

	// Compare Method field
	if reflect.DeepEqual(p.Method, other.Method) {
		return false
	}

	// Compare Insecure field
	if reflect.DeepEqual(p.Insecure, other.Insecure) {
		return false
	}

	// Compare RequestBody field
	if reflect.DeepEqual(p.RequestBody, other.RequestBody) {
		return false
	}

	// Compare RequestHeaders field using deep equality
	if !reflect.DeepEqual(p.RequestHeaders, other.RequestHeaders) {
		return false
	}

	// All fields are equal
	return true

}

func (p *PipelineStepHttp) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	inputs, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, err
	}

	var urlInput string
	if p.UnresolvedAttributes[schema.AttributeTypeUrl] == nil {
		if p.Url == nil {
			return nil, perr.InternalWithMessage("Url must be supplied")
		}
		urlInput = *p.Url
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeUrl], evalContext, &urlInput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}
	inputs[schema.AttributeTypeUrl] = urlInput

	if p.UnresolvedAttributes[schema.AttributeTypeMethod] == nil {
		if p.Method != nil {
			inputs[schema.AttributeTypeMethod] = *p.Method
		}
	} else {
		var method string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMethod], evalContext, &method)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeMethod] = strings.ToLower(method)
	}

	if p.UnresolvedAttributes[schema.AttributeTypeCaCertPem] == nil {
		if p.CaCertPem != nil {
			inputs[schema.AttributeTypeCaCertPem] = *p.CaCertPem
		}
	} else {
		var caCertPem string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCaCertPem], evalContext, &caCertPem)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeCaCertPem] = caCertPem
	}

	if p.UnresolvedAttributes[schema.AttributeTypeInsecure] == nil {
		if p.Insecure != nil {
			inputs[schema.AttributeTypeInsecure] = *p.Insecure
		}
	} else {
		var insecure bool
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeInsecure], evalContext, &insecure)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeInsecure] = insecure
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestBody] == nil {
		if p.RequestBody != nil {
			inputs[schema.AttributeTypeRequestBody] = *p.RequestBody
		}
	} else {
		var requestBody string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestBody], evalContext, &requestBody)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeRequestBody] = requestBody
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestHeaders] == nil {
		if p.RequestHeaders != nil {
			inputs[schema.AttributeTypeRequestHeaders] = p.RequestHeaders
		}
	} else {
		var requestHeaders map[string]string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestHeaders], evalContext, &requestHeaders)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
		inputs[schema.AttributeTypeRequestHeaders] = requestHeaders
	}

	if p.BasicAuthConfig != nil {
		basicAuth, diags := p.BasicAuthConfig.GetInputs(evalContext, p.UnresolvedAttributes)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
		basicAuthMap := make(map[string]interface{})
		basicAuthMap["Username"] = basicAuth.Username
		basicAuthMap["Password"] = basicAuth.Password
		inputs[schema.BlockTypePipelineBasicAuth] = basicAuthMap
	}
	inputs[schema.AttributeTypeStepName] = p.Name

	return inputs, nil
}

func (p *PipelineStepHttp) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeUrl:
			fieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, fieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeMethod:
			fieldName := utils.CapitalizeFirst(name)

			stepDiags := setStringAttribute(attr, evalContext, p, fieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if types.SafeString(p.Method) == "" {
				if !helpers.StringSliceContains(ValidHttpMethods, strings.ToLower(types.SafeString(p.Method))) {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid HTTP method: " + types.SafeString(p.Method),
						Subject:  &attr.Range,
					})
					continue
				}
			}

		case schema.AttributeTypeCaCertPem:
			stepDiags := setStringAttribute(attr, evalContext, p, "CaCertPem", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeInsecure:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				if val.Type() != cty.Bool {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid value for insecure attribute",
						Subject:  &attr.Range,
					})
					continue
				}
				insecure := val.True()
				p.Insecure = &insecure
			}

		case schema.AttributeTypeRequestBody:
			stepDiags := setStringAttribute(attr, evalContext, p, "RequestBody", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeRequestHeaders:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)

			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.RequestHeaders, err = hclhelpers.CtyToGoMapInterface(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse request_headers attribute",
						Subject:  &attr.Range,
					})
					continue
				}
			}
		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for HTTP Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepHttp) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.PipelineStepBase.SetBlockConfig(blocks, evalContext)

	basicAuthConfig := &BasicAuthConfig{}

	if basicAuthBlocks := blocks.ByType()[schema.BlockTypePipelineBasicAuth]; len(basicAuthBlocks) > 0 {
		if len(basicAuthBlocks) > 1 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Multiple basic_auth blocks found for step http",
			}}
		}
		basicAuthBlock := basicAuthBlocks[0]

		var attributes hcl.Attributes

		attributes, diags = basicAuthBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		if attr, exists := attributes[schema.AttributeTypeUsername]; exists {
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				username, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeUsername + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				basicAuthConfig.Username = username
			}

		}

		if attr, exists := attributes[schema.AttributeTypePassword]; exists {
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				password, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypePassword + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				basicAuthConfig.Password = password
			}

		}
		p.BasicAuthConfig = basicAuthConfig
	}

	return diags
}

func (p *PipelineStepHttp) Validate() hcl.Diagnostics {
	diags := p.ValidateBaseAttributes()
	return diags
}
