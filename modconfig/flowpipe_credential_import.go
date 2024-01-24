package modconfig

import (
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
)

// The definition of a single Flowpipe Trigger
type CredentialImport struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	FileName        string `json:"file_name"`
	StartLineNumber int    `json:"start_line_number"`
	EndLineNumber   int    `json:"end_line_number"`

	Source      *string  `json:"source"`
	Connections []string `json:"connections"`
	Prefix      *string  `json:"prefix"`
}

func (c *CredentialImport) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	c.FileName = fileName
	c.StartLineNumber = startLineNumber
	c.EndLineNumber = endLineNumber
}

func (c *CredentialImport) Equals(other *CredentialImport) bool {
	return c.FullName == other.FullName &&
		c.GetMetadata().ModFullName == other.GetMetadata().ModFullName
}

func (c *CredentialImport) GetSource() *string {
	return c.Source
}

func (c *CredentialImport) GetPrefix() *string {
	return c.Prefix
}

func (c *CredentialImport) GetConnections() []string {
	return c.Connections
}

func (c *CredentialImport) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseTriggerAttributes, name)
}

func (c *CredentialImport) SetBaseAttributes(mod *Mod, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			c.Description = desc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTitle]; exists {
		title, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			c.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDocumentation]; exists {
		doc, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			c.Documentation = doc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTags]; exists {
		tags, moreDiags := hclhelpers.AttributeToMap(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			resultMap := make(map[string]string)
			for key, value := range tags {
				resultMap[key] = value.(string)
			}
			c.Tags = resultMap
		}
	}

	return diags
}

func (c *CredentialImport) SetAttributes(mod *Mod, credentialImport *CredentialImport, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := credentialImport.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSource:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			source, err := hclhelpers.CtyToString(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + schema.AttributeTypeSource + " attribute to string",
					Subject:  &attr.Range,
				})
			}
			c.Source = &source
		default:
			if !credentialImport.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Schedule: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

func NewCredentialImport(block *hcl.Block) *CredentialImport {

	credentialImportName := block.Labels[0]

	return &CredentialImport{
		HclResourceImpl: HclResourceImpl{
			FullName:        credentialImportName,
			ShortName:       credentialImportName,
			UnqualifiedName: credentialImportName,
			DeclRange:       block.DefRange,
			blockType:       block.Type,
		},
	}
}
