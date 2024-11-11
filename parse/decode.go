package parse

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

// A consistent detail message for all "not a valid identifier" diagnostics.
const badIdentifierDetail = "A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes."

var missingVariableErrors = []string{
	// returned when the context variables does not have top level 'type' node (locals/control/etc)
	"Unknown variable",
	// returned when the variables have the type object but a field has not yet been populated
	"Unsupported attribute",
	"Missing map element",
}

//
//func decode(parseCtx *ModParseContext) hcl.Diagnostics {
//	utils.LogTime(fmt.Sprintf("decode %s start", parseCtx.CurrentMod.Name()))
//	defer utils.LogTime(fmt.Sprintf("decode %s end", parseCtx.CurrentMod.Name()))
//
//	var diags hcl.Diagnostics
//
//	blocks, err := parseCtx.BlocksToDecode()
//	// build list of blocks to decode
//	if err != nil {
//		diags = append(diags, &hcl.Diagnostic{
//			Severity: hcl.DiagError,
//			Summary:  "failed to determine required dependency order",
//			Detail:   err.Error()})
//		return diags
//	}
//
//	// now clear dependencies from run context - they will be rebuilt
//	parseCtx.ClearDependencies()
//
//	for _, block := range blocks {
//		if block.Type == schema.BlockTypeLocals {
//			resources, res := decodeLocalsBlock(block, parseCtx)
//			if !res.Success() {
//				diags = append(diags, res.Diags...)
//				continue
//			}
//			for _, resource := range resources {
//				resourceDiags := AddResourceToMod(resource, block, parseCtx)
//				diags = append(diags, resourceDiags...)
//			}
//		} else {
//			resource, res := decodeBlock(block, parseCtx)
//			diags = append(diags, res.Diags...)
//			if !res.Success() || resource == nil {
//				continue
//			}
//
//			resourceDiags := AddResourceToMod(resource, block, parseCtx)
//			diags = append(diags, resourceDiags...)
//		}
//	}
//
//	return diags
//}

func AddResourceToMod(resource modconfig.HclResource, block *hcl.Block, decoder Decoder, parseCtx *ModParseContext) hcl.Diagnostics {
	if !decoder.ShouldAddToMod(resource, block, parseCtx) {
		return nil
	}
	return parseCtx.CurrentMod.AddResource(resource)

}

func decodeMod(block *hcl.Block, evalCtx *hcl.EvalContext, mod *modconfig.Mod) (*modconfig.Mod, *DecodeResult) {
	res := NewDecodeResult()

	// decode the database attribute separately
	// do a partial decode using a schema containing just database - use to pull out all other body content in the remain block
	databaseContent, remain, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: schema.AttributeTypeDatabase},
		}})
	res.HandleDecodeDiags(diags)

	// decode the body
	moreDiags := DecodeHclBody(remain, evalCtx, mod, mod)
	res.HandleDecodeDiags(moreDiags)

	connectionStringProvider, searchPath, searchPathPrefix, moreDiags := ResolveConnectionString(databaseContent, evalCtx)
	res.HandleDecodeDiags(moreDiags)

	// if connection string or search path was specified (by the mod referencing a connection), set them
	if connectionStringProvider != nil {
		mod.SetDatabase(connectionStringProvider)
	}
	if searchPath != nil {
		mod.SetSearchPath(searchPath)
	}
	if searchPathPrefix != nil {
		mod.SetSearchPathPrefix(searchPathPrefix)

	}

	return mod, res

}

//func DecodeRequire(block *hcl.Block, evalCtx *hcl.EvalContext) (*modconfig.Require, hcl.Diagnostics) {
//	require := modconfig.NewRequire()
//	// set ranges
//	require.DeclRange = hclhelpers.BlockRange(block)
//	require.TypeRange = block.TypeRange
//	// decode
//	diags := gohcl.DecodeBody(block.Body, evalCtx, require)
//	return require, diags
//}

func ResolveConnectionString(content *hcl.BodyContent, evalCtx *hcl.EvalContext) (csp connection.ConnectionStringProvider, searchPath, searchPathPrefix []string, diags hcl.Diagnostics) {

	attr, exists := content.Attributes[schema.AttributeTypeDatabase]
	if !exists {
		return nil, searchPath, searchPathPrefix, diags
	}

	var dbValue cty.Value
	diags = gohcl.DecodeExpression(attr.Expr, evalCtx, &dbValue)

	if diags.HasErrors() {
		// use decode result to handle any dependencies
		res := NewDecodeResult()
		res.HandleDecodeDiags(diags)
		diags = res.Diags
		// if there are other errors, return them
		if diags.HasErrors() {
			return nil, searchPath, searchPathPrefix, res.Diags
		}
		// so there is a dependency error - if it is for a connection, return the connection name as the connection string
		for _, dep := range res.Depends {
			for _, traversal := range dep.Traversals {
				depName := hclhelpers.TraversalAsString(traversal)
				csp = connection.NewConnectionString(depName)
				if strings.HasPrefix(depName, "connection.") {
					return csp, searchPath, searchPathPrefix, diags
				}
			}
		}
		// if we get here, there is a dependency error but it is not for a connection
		// return the original diags for the calling code to handle
		return nil, searchPath, searchPathPrefix, diags
	}
	// check if this is a connection string or a connection
	if dbValue.Type() == cty.String {
		csp = connection.NewConnectionString(dbValue.AsString())
	} else {
		// if this is a temporary connection, ignore (this will only occur during the variable parsing phase)
		if dbValue.Type().HasAttribute("temporary") {
			return nil, searchPath, searchPathPrefix, diags
		}

		c, err := app_specific_connection.CtyValueToConnection(dbValue)
		if err != nil {
			return nil, searchPath, searchPathPrefix, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  err.Error(),
					Subject:  attr.Range.Ptr(),
				}}
		}

		var ok bool
		csp, ok = c.(connection.ConnectionStringProvider)
		if !ok {
			// the connection type must support connection strings

			slog.Warn("connection does not support connection string", "db", c)
			return nil, searchPath, searchPathPrefix, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "invalid connection reference - only connections which implement GetConnectionString() are supported",
				}}
		}
		if conn, ok := c.(connection.SearchPathProvider); ok {
			searchPath = conn.GetSearchPath()
			searchPathPrefix = conn.GetSearchPathPrefix()
		}
	}

	return csp, searchPath, searchPathPrefix, diags
}

func DecodeProperty(content *hcl.BodyContent, property string, dest interface{}, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if attr, ok := content.Attributes[property]; ok {
		diags = gohcl.DecodeExpression(attr.Expr, evalCtx, dest)
	}
	return diags
}

func resourceIsAnonymous(resource modconfig.HclResource) bool {
	// (if a resource anonymous it must support ResourceWithMetadata)
	resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata)
	anonymousResource := ok && resourceWithMetadata.IsAnonymous()
	return anonymousResource
}

func AddResourceMetadata(resourceWithMetadata modconfig.ResourceWithMetadata, srcRange hcl.Range, parseCtx *ModParseContext) hcl.Diagnostics {
	metadata, err := GetMetadataForParsedResource(resourceWithMetadata.Name(), srcRange, parseCtx.FileData, parseCtx.CurrentMod)
	if err != nil {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
			Subject:  &srcRange,
		}}
	}
	//  set on resource
	resourceWithMetadata.SetMetadata(metadata)
	return nil
}

func ValidateName(block *hcl.Block) hcl.Diagnostics {
	if len(block.Labels) == 0 {
		return nil
	}

	if !hclsyntax.ValidIdentifier(block.Labels[0]) {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		}}
	}
	return nil
}

// Validate all blocks and attributes are supported
// We use partial decoding so that we can automatically decode as many properties as possible
// and only manually decode properties requiring special logic.
// The problem is the partial decode does not return errors for invalid attributes/blocks, so we must implement our own
func validateHcl(blockType string, body *hclsyntax.Body, schema *hcl.BodySchema) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// identify any blocks specified by hcl tags
	var supportedBlocks = make(map[string]struct{})
	var supportedAttributes = make(map[string]struct{})
	for _, b := range schema.Blocks {
		supportedBlocks[b.Type] = struct{}{}
	}
	for _, b := range schema.Attributes {
		supportedAttributes[b.Name] = struct{}{}
	}

	// now check for invalid blocks
	for _, block := range body.Blocks {
		if _, ok := supportedBlocks[block.Type]; !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf(`Unsupported block type: Blocks of type '%s' are not expected here.`, block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}
	for _, attribute := range body.Attributes {
		if _, ok := supportedAttributes[attribute.Name]; !ok {
			// special case code for deprecated properties
			subject := attribute.Range()
			if isDeprecated(attribute, blockType) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  fmt.Sprintf(`Deprecated attribute: '%s' is deprecated for '%s' blocks and will be ignored.`, attribute.Name, blockType),
					Subject:  &subject,
				})
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf(`Unsupported attribute: '%s' not expected here.`, attribute.Name),
					Subject:  &subject,
				})
			}
		}
	}

	return diags
}

func isDeprecated(attribute *hclsyntax.Attribute, blockType string) bool {
	switch attribute.Name {
	case "search_path", "search_path_prefix":
		return blockType == schema.BlockTypeQuery || blockType == schema.BlockTypeControl
	default:
		return false
	}
}
