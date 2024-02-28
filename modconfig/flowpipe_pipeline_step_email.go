package modconfig

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

// Deprecated: Pipeline step email is deprecated. Please use the new pipeline message step with notifier.
type PipelineStepEmail struct {
	PipelineStepBase
	To           []string `json:"to"`
	From         *string  `json:"from"`
	SmtpPassword *string  `json:"smtp_password"`
	SmtpUsername *string  `json:"smtp_username"`
	Host         *string  `json:"host"`
	Port         *int64   `json:"port"`
	SenderName   *string  `json:"sender_name"`
	Cc           []string `json:"cc"`
	Bcc          []string `json:"bcc"`
	Body         *string  `json:"body"`
	ContentType  *string  `json:"content_type"`
	Subject      *string  `json:"subject"`
}

func (p *PipelineStepEmail) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepEmail)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Use reflect.DeepEqual to compare slices and pointers
	return reflect.DeepEqual(p.To, other.To) &&
		reflect.DeepEqual(p.From, other.From) &&
		reflect.DeepEqual(p.SmtpUsername, other.SmtpUsername) &&
		reflect.DeepEqual(p.SmtpPassword, other.SmtpPassword) &&
		reflect.DeepEqual(p.Host, other.Host) &&
		reflect.DeepEqual(p.Port, other.Port) &&
		reflect.DeepEqual(p.SenderName, other.SenderName) &&
		reflect.DeepEqual(p.Cc, other.Cc) &&
		reflect.DeepEqual(p.Bcc, other.Bcc) &&
		reflect.DeepEqual(p.Body, other.Body) &&
		reflect.DeepEqual(p.ContentType, other.ContentType) &&
		reflect.DeepEqual(p.Subject, other.Subject)

}

func (p *PipelineStepEmail) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var to []string
	if p.UnresolvedAttributes[schema.AttributeTypeTo] == nil {
		to = p.To
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeTo], evalContext, &to)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var from *string
	if p.UnresolvedAttributes[schema.AttributeTypeFrom] == nil {
		from = p.From
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeFrom], evalContext, &from)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var smtpUsername *string
	if p.UnresolvedAttributes[schema.AttributeTypeSmtpUsername] == nil {
		smtpUsername = p.SmtpUsername
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSmtpUsername], evalContext, &smtpUsername)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var smtpPassword *string
	if p.UnresolvedAttributes[schema.AttributeTypeSmtpPassword] == nil {
		smtpPassword = p.SmtpPassword
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSmtpPassword], evalContext, &smtpPassword)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var host *string
	if p.UnresolvedAttributes[schema.AttributeTypeHost] == nil {
		host = p.Host
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeHost], evalContext, &host)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var port *int64
	if p.UnresolvedAttributes[schema.AttributeTypePort] == nil {
		port = p.Port
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypePort], evalContext, &port)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var senderName *string
	if p.UnresolvedAttributes[schema.AttributeTypeSenderName] == nil {
		senderName = p.SenderName
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSenderName], evalContext, &senderName)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var body *string
	if p.UnresolvedAttributes[schema.AttributeTypeBody] == nil {
		body = p.Body
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeBody], evalContext, &body)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var subject *string
	if p.UnresolvedAttributes[schema.AttributeTypeSubject] == nil {
		subject = p.Subject
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSubject], evalContext, &subject)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var contentType *string
	if p.UnresolvedAttributes[schema.AttributeTypeContentType] == nil {
		contentType = p.ContentType
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeContentType], evalContext, &contentType)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var cc []string
	if p.UnresolvedAttributes[schema.AttributeTypeCc] == nil {
		cc = p.Cc
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCc], evalContext, &cc)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	var bcc []string
	if p.UnresolvedAttributes[schema.AttributeTypeBcc] == nil {
		bcc = p.Bcc
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeBcc], evalContext, &bcc)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(p.Name, diags)
		}
	}

	results := map[string]interface{}{}

	if to != nil {
		results[schema.AttributeTypeTo] = to
	}

	if from != nil {
		results[schema.AttributeTypeFrom] = *from
	}

	if smtpUsername != nil {
		results[schema.AttributeTypeSmtpUsername] = *smtpUsername
	}

	if smtpPassword != nil {
		results[schema.AttributeTypeSmtpPassword] = *smtpPassword
	}

	if host != nil {
		results[schema.AttributeTypeHost] = *host
	}

	if port != nil {
		results[schema.AttributeTypePort] = *port
	}

	if senderName != nil {
		results[schema.AttributeTypeSenderName] = *senderName
	}

	if cc != nil {
		results[schema.AttributeTypeCc] = cc
	}

	if bcc != nil {
		results[schema.AttributeTypeBcc] = bcc
	}

	if body != nil {
		results[schema.AttributeTypeBody] = *body
	}

	if contentType != nil {
		results[schema.AttributeTypeContentType] = *contentType
	}

	if subject != nil {
		results[schema.AttributeTypeSubject] = *subject
	}

	return results, nil
}

func (p *PipelineStepEmail) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeTo:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				emailRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeTo + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.To = emailRecipients
			}

		case schema.AttributeTypeFrom:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				from, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeFrom + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.From = &from
			}

		case schema.AttributeTypeSmtpUsername:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				smtpUsername, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSmtpUsername + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SmtpUsername = &smtpUsername
			}

		case schema.AttributeTypeSmtpPassword:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				smtpPassword, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSmtpPassword + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SmtpPassword = &smtpPassword
			}

		case schema.AttributeTypeHost:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				host, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeHost + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Host = &host
			}

		case schema.AttributeTypePort:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				port, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to convert port into integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Port = port
			}

		case schema.AttributeTypeSenderName:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				senderName, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSenderName + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SenderName = &senderName
			}

		case schema.AttributeTypeCc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				ccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeCc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Cc = ccRecipients
			}

		case schema.AttributeTypeBcc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				bccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBcc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Bcc = bccRecipients
			}

		case schema.AttributeTypeBody:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				body, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBody + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Body = &body
			}

		case schema.AttributeTypeContentType:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				contentType, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeContentType + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.ContentType = &contentType
			}

		case schema.AttributeTypeSubject:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				subject, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSubject + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Subject = &subject
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Email Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}
