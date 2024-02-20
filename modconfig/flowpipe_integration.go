package modconfig

import (
	"fmt"

	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Integration interface {
	HclResource
	ResourceWithMetadata

	CtyValue() (cty.Value, error)
	GetIntegrationImpl() *IntegrationImpl
	GetIntegrationType() string
	MapInterface() (map[string]interface{}, error)
	SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics
	SetFileReference(fileName string, startLineNumber int, endLineNumber int)
	SetUrl(string)
	Validate() hcl.Diagnostics
}

type IntegrationImpl struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	// Slack and Webform has URL, Email integration does not it will be null
	Url             *string `json:"url,omitempty" cty:"url" hcl:"url,optional"`
	FileName        string
	StartLineNumber int
	EndLineNumber   int
}

func (i *IntegrationImpl) SetUrl(url string) {
	i.Url = &url
}

func (i *IntegrationImpl) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	i.FileName = fileName
	i.StartLineNumber = startLineNumber
	i.EndLineNumber = endLineNumber
}

func (i *IntegrationImpl) GetIntegrationImpl() *IntegrationImpl {
	return i
}

func DefaultIntegrations() (map[string]Integration, error) {
	integrations := make(map[string]Integration)

	webhookIntegration := &WebformIntegration{
		HclResourceImpl: HclResourceImpl{
			FullName:        schema.IntegrationTypeWebform + ".default",
			ShortName:       "default",
			UnqualifiedName: schema.IntegrationTypeWebform + ".default",
		},
		Type: schema.IntegrationTypeWebform,
	}

	integrations[schema.IntegrationTypeWebform+".default"] = webhookIntegration

	return integrations, nil
}

type SlackIntegration struct {
	HclResourceImpl          `json:"-"`
	ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl          `json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`

	// slack
	Token         *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
	SigningSecret *string `json:"signing_secret,omitempty" cty:"signing_secret" hcl:"signing_secret,optional"`
	WebhookUrl    *string `json:"webhook_url,omitempty" cty:"webhook_url" hcl:"webhook_url,optional"`
	Channel       *string `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
}

func (i *SlackIntegration) CtyValue() (cty.Value, error) {
	return GetCtyValue(i)
}

func (i *SlackIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["type"] = i.Type
	if i.Token != nil {
		res["token"] = *i.Token
	}
	if i.SigningSecret != nil {
		res["signing_secret"] = *i.SigningSecret
	}
	if i.WebhookUrl != nil {
		res["webhook_url"] = *i.WebhookUrl
	}
	if i.Channel != nil {
		res["channel"] = *i.Channel
	}
	return res, nil
}

func (i *SlackIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *SlackIntegration) Validate() hcl.Diagnostics {
	// TODO: slack integration validation
	return hcl.Diagnostics{}
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

	if (i.Channel == nil && other.Channel != nil) || (i.Channel != nil && other.Channel == nil) || (i.Channel != nil && other.Channel != nil && *i.Channel != *other.Channel) {
		return false
	}

	// If all fields are equal, the structs are equal
	return true
}

func (i *SlackIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics
	var whSet, tknSet bool

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeToken:
			token, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.Token = token
			tknSet = true
		case schema.AttributeTypeSigningSecret:
			ss, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SigningSecret = ss
		case schema.AttributeTypeWebhookUrl:
			webhookUrl, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.WebhookUrl = webhookUrl
			whSet = true
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Slack Integration: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	if tknSet && whSet {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attributes token and webhook_url are mutually exclusive: " + i.Name(),
		})
	}

	if !tknSet && !whSet {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  i.Name() + " requires one of the following attributes set: token, webhook_url",
		})
	}

	return diags
}

type EmailIntegration struct {
	HclResourceImpl          `json:"-"`
	ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl          `json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`

	// email
	SmtpHost     *string `json:"smtp_host,omitempty" cty:"smtp_host" hcl:"smtp_host"`
	SmtpTls      *string `json:"smtp_tls,omitempty" cty:"smtp_tls" hcl:"smtp_tls,optional"`
	SmtpPort     *int    `json:"smtp_port,omitempty" cty:"smtp_port" hcl:"smtp_port,optional"`
	SmtpsPort    *int    `json:"smtps_port,omitempty" cty:"smtps_port" hcl:"smtps_port,optional"`
	SmtpUsername *string `json:"smtp_username,omitempty" cty:"smtp_username" hcl:"smtp_username,optional"`
	SmtpPassword *string `json:"smtp_password,omitempty" cty:"smtp_password" hcl:"smtp_password,optional"`

	From             *string `json:"from,omitempty" cty:"from" hcl:"from"`
	DefaultRecipient *string `json:"default_recipient,omitempty" cty:"default_recipient" hcl:"default_recipient,optional"`
	DefaultSubject   *string `json:"default_subject,omitempty" cty:"default_subject" hcl:"default_subject,optional"`
	ResponseUrl      *string `json:"response_url,omitempty" cty:"response_url" hcl:"response_url,optional"`
}

func (i *EmailIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["type"] = i.Type
	if i.SmtpHost != nil {
		res["smtp_host"] = *i.SmtpHost
	}
	if i.SmtpTls != nil {
		res["smtp_tls"] = *i.SmtpTls
	}
	if i.SmtpPort != nil {
		res["smtp_port"] = *i.SmtpPort
	}
	if i.SmtpsPort != nil {
		res["smtps_port"] = *i.SmtpsPort
	}
	if i.SmtpUsername != nil {
		res["smtp_username"] = *i.SmtpUsername
	}
	if i.SmtpPassword != nil {
		res["smtp_password"] = *i.SmtpPassword
	}
	if i.From != nil {
		res["from"] = *i.From
	}
	if i.DefaultRecipient != nil {
		res["default_recipient"] = *i.DefaultRecipient
	}
	if i.DefaultSubject != nil {
		res["default_subject"] = *i.DefaultSubject
	}
	if i.ResponseUrl != nil {
		res["response_url"] = *i.ResponseUrl
	}
	return res, nil
}

func (i *EmailIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *EmailIntegration) CtyValue() (cty.Value, error) {
	return GetCtyValue(i)
}

func (i *EmailIntegration) Validate() hcl.Diagnostics {
	// TODO: email integration validation
	return hcl.Diagnostics{}
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

func (i *EmailIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSmtpHost:
			host, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpHost = host
		case schema.AttributeTypeSmtpTls:
			tls, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpTls = tls
		case schema.AttributeTypeSmtpPort:
			port, moreDiags := hclhelpers.AttributeToInt(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			portInt := int(*port)
			i.SmtpPort = &portInt
		case schema.AttributeTypeSmtpsPort:
			port, moreDiags := hclhelpers.AttributeToInt(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			portInt := int(*port)
			i.SmtpsPort = &portInt
		case schema.AttributeTypeSmtpUsername:
			uName, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpUsername = uName
		case schema.AttributeTypeSmtpPassword:
			pass, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.SmtpPassword = pass
		case schema.AttributeTypeFrom:
			from, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.From = from
		case schema.AttributeTypeDefaultRecipient:
			rec, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.DefaultRecipient = rec
		case schema.AttributeTypeDefaultSubject:
			subject, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.DefaultSubject = subject
		case schema.AttributeTypeResponseUrl:
			url, moreDiags := hclhelpers.AttributeToString(attr, evalContext, false)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}
			i.ResponseUrl = url
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Email Integration: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

func integrationFromCtyValue(val cty.Value) (Integration, error) {

	if val.IsNull() || val == cty.NilVal {
		return nil, perr.BadRequestWithMessage("Integration is required")
	}

	valMap := val.AsValueMap()
	integrationType := valMap["type"]

	if integrationType.IsNull() || integrationType == cty.NilVal {
		return nil, perr.BadRequestWithMessage("Integration type is required")
	}

	switch integrationType.AsString() {
	case schema.IntegrationTypeSlack:
		return SlackIntegrationFromCtyValue(val)
	case schema.IntegrationTypeEmail:
		return EmailIntegrationFromCtyValue(val)
	case schema.IntegrationTypeWebform:
		return WebformIntegrationFromCtyValue(val)
	}
	return nil, perr.BadRequestWithMessage(fmt.Sprintf("Unsupported integration type: %s", integrationType))
}

func SlackIntegrationFromCtyValue(val cty.Value) (*SlackIntegration, error) {
	i := &SlackIntegration{}

	i.Type = val.GetAttr("type").AsString()

	valMap := val.AsValueMap()

	token := valMap["token"]
	signingSecret := valMap["signing_secret"]
	webhookUrl := valMap["webhook_url"]
	channel := valMap["channel"]

	if !token.IsNull() {
		tokenStr := token.AsString()
		i.Token = &tokenStr
	}

	if !signingSecret.IsNull() {
		signingSecretStr := signingSecret.AsString()
		i.SigningSecret = &signingSecretStr
	}

	if !webhookUrl.IsNull() {
		webhookUrlStr := webhookUrl.AsString()
		i.WebhookUrl = &webhookUrlStr
	}

	if !channel.IsNull() {
		channelStr := channel.AsString()
		i.Channel = &channelStr
	}

	return i, nil
}

func EmailIntegrationFromCtyValue(val cty.Value) (*EmailIntegration, error) {
	i := &EmailIntegration{}

	i.Type = val.GetAttr("type").AsString()

	valMap := val.AsValueMap()

	smtpHost := valMap["smtp_host"]
	smtpTls := valMap["smtp_tls"]
	smtpPort := valMap["smtp_port"]
	smtpsPort := valMap["smtps_port"]
	smtpUsername := valMap["smtp_username"]
	smtpPassword := valMap["smtp_password"]
	from := valMap["from"]
	defaultRecipient := valMap["default_recipient"]
	defaultSubject := valMap["default_subject"]
	responseUrl := valMap["response_url"]

	if !smtpHost.IsNull() {
		smtpHostStr := smtpHost.AsString()
		i.SmtpHost = &smtpHostStr
	}

	if !smtpTls.IsNull() {
		smtpTlsStr := smtpTls.AsString()
		i.SmtpTls = &smtpTlsStr
	}

	if !smtpPort.IsNull() {
		smtpPortInt, _ := smtpPort.AsBigFloat().Int64()
		n := int(smtpPortInt)
		i.SmtpPort = &n
	}

	if !smtpsPort.IsNull() {
		smtpsPortInt, _ := smtpsPort.AsBigFloat().Int64()
		n := int(smtpsPortInt)
		i.SmtpsPort = &n
	}

	if !smtpUsername.IsNull() {
		smtpUsernameStr := smtpUsername.AsString()
		i.SmtpUsername = &smtpUsernameStr
	}

	if !smtpPassword.IsNull() {
		smtpPasswordStr := smtpPassword.AsString()
		i.SmtpPassword = &smtpPasswordStr
	}

	if !from.IsNull() {
		fromStr := from.AsString()
		i.From = &fromStr
	}

	if !defaultRecipient.IsNull() {
		defaultRecipientStr := defaultRecipient.AsString()
		i.DefaultRecipient = &defaultRecipientStr
	}

	if !defaultSubject.IsNull() {
		defaultSubjectStr := defaultSubject.AsString()
		i.DefaultSubject = &defaultSubjectStr
	}

	if !responseUrl.IsNull() {
		responseUrlStr := responseUrl.AsString()
		i.ResponseUrl = &responseUrlStr
	}

	return i, nil
}

func WebformIntegrationFromCtyValue(val cty.Value) (*WebformIntegration, error) {
	i := &WebformIntegration{}

	i.Type = val.GetAttr("type").AsString()

	return i, nil
}

func NewIntegrationFromBlock(block *hcl.Block) Integration {
	integrationType := block.Labels[0]
	integrationName := block.Labels[1]

	integrationFullName := integrationType + "." + integrationName

	hclResourceImpl := HclResourceImpl{
		FullName:        integrationFullName,
		UnqualifiedName: integrationFullName,
		ShortName:       integrationName,
		DeclRange:       block.DefRange,
		blockType:       block.Type,
	}

	switch integrationType {
	case schema.IntegrationTypeSlack:
		return &SlackIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
		}
	case schema.IntegrationTypeEmail:
		return &EmailIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
		}
	case schema.IntegrationTypeWebform:
		return &WebformIntegration{
			HclResourceImpl: hclResourceImpl,
			Type:            integrationType,
		}
	}

	return nil
}

type WebformIntegration struct {
	HclResourceImpl          `json:"-"`
	ResourceWithMetadataImpl `json:"-"`
	IntegrationImpl          `json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (i *WebformIntegration) GetIntegrationType() string {
	return i.Type
}

func (i *WebformIntegration) CtyValue() (cty.Value, error) {
	return GetCtyValue(i)
}

func (i *WebformIntegration) MapInterface() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["type"] = i.Type
	return res, nil
}

func (i *WebformIntegration) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (i *WebformIntegration) Equals(other *WebformIntegration) bool {
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

	// Compare the fields specific to WebformIntegration
	if i.Type != other.Type {
		return false
	}

	// If all fields are equal, the structs are equal
	return true
}

func (i *WebformIntegration) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	return hcl.Diagnostics{}
}
