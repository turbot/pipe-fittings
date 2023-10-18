package modconfig

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
)

type Integration struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	IntegrationType string

	// mod *Mod

	Config IIntegrationConfig `json:"-"`
}

func (i *Integration) Equals(other *Integration) bool {
	if i == nil && other == nil {
		return true
	}
	if i == nil || other == nil {
		return false
	}
	if i.IntegrationType != other.IntegrationType {
		return false
	}

	// i.Config should never be null
	return i.Config.Equals(other.Config) && i.Name() == other.Name()
}

type IIntegrationConfig interface {
	SetAttributes(*Mod, *Integration, hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
	Equals(other IIntegrationConfig) bool
}

type IntegrationConfigSlack struct {
	SigningSecret *string `json:"signing_secret,omitempty"`
	Token         *string `json:"token,omitempty"`
	WebhookUrl    *string `json:"webhook_url,omitempty"`
}

func (i *IntegrationConfigSlack) SetAttributes(mod *Mod, integration *Integration, attributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attr := range attributes {
		switch name {
		case schema.AttributeTypeSigningSecret:
			signingSecret, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SigningSecret = signingSecret

		case schema.AttributeTypeToken:
			token, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			i.Token = token

		case schema.AttributeTypeWebhookUrl:
			webhookUrl, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			i.WebhookUrl = webhookUrl

		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unknown attribute",
				Detail:   "Unknown attribute " + name + " for " + integration.Name(),
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

func (i *IntegrationConfigSlack) Equals(other IIntegrationConfig) bool {
	if i == nil && other == nil {
		return true
	}
	if i == nil || other == nil {
		return false
	}

	otherSlack, ok := other.(*IntegrationConfigSlack)
	if !ok {
		return false
	}
	if i.SigningSecret != nil && otherSlack.SigningSecret != nil && *i.SigningSecret != *otherSlack.SigningSecret {
		return false
	}
	if i.Token != nil && otherSlack.Token != nil && *i.Token != *otherSlack.Token {
		return false
	}
	if i.WebhookUrl != nil && otherSlack.WebhookUrl != nil && *i.WebhookUrl != *otherSlack.WebhookUrl {
		return false
	}
	return true
}

func NewIntegration(mod *Mod, block *hcl.Block) *Integration {

	integrationType := block.Labels[0]
	integrationName := block.Labels[1]

	// TODO: rethink this area, we need to be able to handle pipelines that are not in a mod
	// TODO: we're trying to integrate the pipeline & trigger functionality into the mod system, so it will look
	// TODO: like a clutch for now
	if mod != nil {
		modName := mod.Name()
		if strings.HasPrefix(modName, "mod") {
			modName = strings.TrimPrefix(modName, "mod.")
		}
		integrationName = modName + ".integration." + integrationName
	} else {
		integrationName = "local.integration." + integrationName
	}

	integration := &Integration{
		HclResourceImpl: HclResourceImpl{
			// The FullName is the full name of the resource, including the mod name
			FullName:        integrationName,
			UnqualifiedName: "pipeline." + block.Labels[0],
			DeclRange:       block.DefRange,
			blockType:       block.Type,
		},
		IntegrationType: integrationType,
	}

	if integrationType == schema.IntegrationTypeSlack {
		integration.Config = &IntegrationConfigSlack{}
	} else {
		return nil
	}

	return integration
}
