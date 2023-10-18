package modconfig

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type IIntegration interface {
	HclResource
	ResourceWithMetadata
	GetType() string
}

type SlackIntegration struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	// slack
	Token         *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	SigningSecret *string `json:"signing_secret,omitempty" cty:"signing_secret" hcl:"signing_secret,optional"`
	WebhookUrl    *string `json:"webhook_url,omitempty" cty:"webhook_url" hcl:"webhook_url,optional"`
}

func (i *SlackIntegration) GetType() string {
	return i.Type
}

func (i *SlackIntegration) Equals(other *SlackIntegration) bool {
	if i == nil && other == nil {
		return true
	}
	if i == nil || other == nil {
		return false
	}

	// Check fields from embedded structs
	if !i.HclResourceImpl.Equals(&other.HclResourceImpl) {
		return false
	}

	// Compare the fields specific to SlackIntegration
	if i.Type != other.Type {
		return false
	}

	if (i.Token == nil && other.Token != nil) || (i.Token != nil && other.Token == nil) || (i.Token != nil && other.Token != nil && *i.Token != *other.Token) {
		return false
	}

	if (i.SigningSecret == nil && other.SigningSecret != nil) || (i.SigningSecret != nil && other.SigningSecret == nil) || (i.SigningSecret != nil && other.SigningSecret != nil && *i.SigningSecret != *other.SigningSecret) {
		return false
	}

	if (i.WebhookUrl == nil && other.WebhookUrl != nil) || (i.WebhookUrl != nil && other.WebhookUrl == nil) || (i.WebhookUrl != nil && other.WebhookUrl != nil && *i.WebhookUrl != *other.WebhookUrl) {
		return false
	}

	// If all fields are equal, the structs are equal
	return true
}

type EmailIntegration struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	// email
	SmtpHost     *string `json:"smtp_host,omitempty" cty:"smtp_host" hcl:"smtp_host,optional"`
	SmtpTls      *string `json:"smtp_tls,omitempty" cty:"smtp_tls" hcl:"smtp_tls,optional"`
	SmtpPort     *int    `json:"smtp_port,omitempty" cty:"smtp_port" hcl:"smtp_port,optional"`
	SmtpsPort    *int    `json:"smtps_port,omitempty" cty:"smtps_port" hcl:"smtps_port,optional"`
	SmtpUsername *string `json:"smtp_username,omitempty" cty:"smtp_username" hcl:"smtp_username,optional"`
	SmtpPassword *string `json:"smtp_password,omitempty" cty:"smtp_password" hcl:"smtp_password,optional"`

	From             *string `json:"from,omitempty" cty:"from" hcl:"from,optional"`
	DefaultRecipient *string `json:"default_recipient,omitempty" cty:"default_recipient" hcl:"default_recipient,optional"`
	DefaultSubject   *string `json:"default_subject,omitempty" cty:"default_subject" hcl:"default_subject,optional"`
}

func (i *EmailIntegration) GetType() string {
	return i.Type
}

func (i *EmailIntegration) Equals(other *EmailIntegration) bool {
	if i == nil && other == nil {
		return true
	}
	if i == nil || other == nil {
		return false
	}

	// Check fields from embedded structs
	if !i.HclResourceImpl.Equals(&other.HclResourceImpl) {
		return false
	}

	// Compare the fields specific to EmailIntegration
	if i.Type != other.Type {
		return false
	}

	if (i.SmtpHost == nil && other.SmtpHost != nil) || (i.SmtpHost != nil && other.SmtpHost == nil) || (i.SmtpHost != nil && other.SmtpHost != nil && *i.SmtpHost != *other.SmtpHost) {
		return false
	}

	if (i.SmtpTls == nil && other.SmtpTls != nil) || (i.SmtpTls != nil && other.SmtpTls == nil) || (i.SmtpTls != nil && other.SmtpTls != nil && *i.SmtpTls != *other.SmtpTls) {
		return false
	}

	if (i.SmtpPort == nil && other.SmtpPort != nil) || (i.SmtpPort != nil && other.SmtpPort == nil) || (i.SmtpPort != nil && other.SmtpPort != nil && *i.SmtpPort != *other.SmtpPort) {
		return false
	}

	if (i.SmtpsPort == nil && other.SmtpsPort != nil) || (i.SmtpsPort != nil && other.SmtpsPort == nil) || (i.SmtpsPort != nil && other.SmtpsPort != nil && *i.SmtpsPort != *other.SmtpsPort) {
		return false
	}

	if (i.SmtpUsername == nil && other.SmtpUsername != nil) || (i.SmtpUsername != nil && other.SmtpUsername == nil) || (i.SmtpUsername != nil && other.SmtpUsername != nil && *i.SmtpUsername != *other.SmtpUsername) {
		return false
	}

	if (i.SmtpPassword == nil && other.SmtpPassword != nil) || (i.SmtpPassword != nil && other.SmtpPassword == nil) || (i.SmtpPassword != nil && other.SmtpPassword != nil && *i.SmtpPassword != *other.SmtpPassword) {
		return false
	}

	if (i.From == nil && other.From != nil) || (i.From != nil && other.From == nil) || (i.From != nil && other.From != nil && *i.From != *other.From) {
		return false
	}

	if (i.DefaultRecipient == nil && other.DefaultRecipient != nil) || (i.DefaultRecipient != nil && other.DefaultRecipient == nil) || (i.DefaultRecipient != nil && other.DefaultRecipient != nil && *i.DefaultRecipient != *other.DefaultRecipient) {
		return false
	}

	if (i.DefaultSubject == nil && other.DefaultSubject != nil) || (i.DefaultSubject != nil && other.DefaultSubject == nil) || (i.DefaultSubject != nil && other.DefaultSubject != nil && *i.DefaultSubject != *other.DefaultSubject) {
		return false
	}

	// If all fields are equal, the structs are equal
	return true
}

func NewIntegration(mod *Mod, block *hcl.Block) IIntegration {

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

	if integrationType == "slack" {
		integration := &SlackIntegration{
			HclResourceImpl: HclResourceImpl{
				// The FullName is the full name of the resource, including the mod name
				FullName:        integrationName,
				UnqualifiedName: "pipeline." + block.Labels[0],
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "slack",
		}
		return integration
	} else if integrationType == "email" {
		integration := &EmailIntegration{
			HclResourceImpl: HclResourceImpl{
				// The FullName is the full name of the resource, including the mod name
				FullName:        integrationName,
				UnqualifiedName: "pipeline." + block.Labels[0],
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "email",
		}
		return integration
	}

	return nil
}
