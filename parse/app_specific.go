package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

// ModDecoderFunc is the appspecific constructor function used to construct a decoder which knows how to
// decode resources for the app in question
var ModDecoderFunc func(...DecoderOption) Decoder

var AppSpecificGetResourceSchemaFunc func(resource modconfig.HclResource, bodySchema *hcl.BodySchema) *hcl.BodySchema
