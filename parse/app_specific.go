package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

// ModDecoderFunc is the appspecific constructor function used to construct a decoder which knows how to
// decode resources for the app in question
var ModDecoderFunc func(...DecoderOption) Decoder

var AppSpecificGetResourceSchemaFunc func(resource modconfig.HclResource, bodySchema *hcl.BodySchema) *hcl.BodySchema

// AppSpecificParseResourceNameFunc provides a mechanism to override the resource name parser used by DecodeHclBody
// it is set by Tailpipe to use its own ParsedResourceName type
var AppSpecificParseResourceNameFunc = func(propertyPath string) (modconfig.ResourceNameProvider, error) {
	return modconfig.ParseResourceName(propertyPath)
}
