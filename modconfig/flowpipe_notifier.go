package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type Notifier struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Notifies []Notify `json:"notifies" cty:"notifies" hcl:"notifies"`
}

func DefaultNotifiers(defaultWebformIntegration Integration) (map[string]Notifier, error) {
	notifiers := make(map[string]Notifier)

	notifier := Notifier{
		HclResourceImpl: HclResourceImpl{
			FullName:        "default",
			ShortName:       "default",
			UnqualifiedName: "default",
		},
	}

	notify := Notify{
		Integration: defaultWebformIntegration,
	}
	notifier.Notifies = []Notify{notify}

	notifiers["default"] = notifier

	return notifiers, nil
}

func (c *Notifier) CtyValue() (cty.Value, error) {
	notifierMap := make(map[string]cty.Value)
	notifies := make([]cty.Value, 0)
	for _, notify := range c.Notifies {
		notifyCtyValue, err := notify.CtyValue()
		if err != nil {
			return cty.NilVal, err
		}
		notifies = append(notifies, notifyCtyValue)
	}

	notifierMap["notifies"] = cty.ListVal(notifies)

	return cty.ObjectVal(notifierMap), nil
}

type Notify struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Integration Integration `json:"integration"`

	Cc          []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc         []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel     *string  `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Description *string  `json:"description,omitempty" cty:"description" hcl:"description,optional"`
	Subject     *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	Title       *string  `json:"title,omitempty" cty:"title" hcl:"title,optional"`
	To          []string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}

func (n *Notify) CtyValue() (cty.Value, error) {
	notifyMap := make(map[string]interface{})

	var err error
	notifyMap["integration"], err = n.Integration.CtyValue()
	if err != nil {
		return cty.NilVal, err
	}

	if n.Cc != nil {
		ccCtys := make([]cty.Value, 0)
		for _, cc := range n.Cc {
			ccCtys = append(ccCtys, cty.StringVal(cc))
		}
		notifyMap["cc"] = cty.ListVal(ccCtys)
	}

	if n.Bcc != nil {
		notifyMap["bcc"] = n.Bcc
		bccCtys := make([]cty.Value, 0)
		for _, bcc := range n.Bcc {
			bccCtys = append(bccCtys, cty.StringVal(bcc))
		}

		notifyMap["bcc"] = cty.ListVal(bccCtys)
	}

	if n.Channel != nil {
		notifyMap["channel"] = cty.StringVal(*n.Channel)
	}

	if n.Description != nil {
		notifyMap["description"] = cty.StringVal(*n.Description)
	}

	if n.Subject != nil {
		notifyMap["subject"] = cty.StringVal(*n.Subject)
	}

	if n.Title != nil {
		notifyMap["title"] = cty.StringVal(*n.Title)
	}

	if n.To != nil {
		toCtys := make([]cty.Value, 0)
		for _, to := range n.To {
			toCtys = append(toCtys, cty.StringVal(to))
		}

		notifyMap["to"] = cty.ListVal(toCtys)
	}

	return cty.ObjectVal(notifyMap), nil
}

func (n *Notify) SetAttributes(body hcl.Body, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	attribs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	attr := attribs[schema.AttributeTypeIntegration]
	if attr == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "integration is required",
				Subject:  body.MissingItemRange().Ptr(),
			},
		}
	}

	integrationCtys, diags := attr.Expr.Value(evalCtx)
	if diags.HasErrors() {
		return diags
	}

	integration, err := integrationFromCtyValue(integrationCtys)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "error decoding integration",
				Subject:  body.MissingItemRange().Ptr(),
			},
		}
	}

	n.Integration = integration
	return diags
}
