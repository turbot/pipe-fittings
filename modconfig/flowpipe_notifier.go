package modconfig

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type Notifier struct {
	HclResourceImpl          `json:"-"`
	ResourceWithMetadataImpl `json:"-"`

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
	notifies := []any{}

	for _, notify := range c.Notifies {
		mapInterface, err := notify.MapInterface()
		if err != nil {
			return cty.NilVal, err
		}

		notifies = append(notifies, mapInterface)
	}

	notifierMap := make(map[string]interface{}, 1)
	notifierMap["notifies"] = notifies

	notifierCtyVal, err := hclhelpers.ConvertInterfaceToCtyValue(notifierMap)
	return notifierCtyVal, err
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

// UnmarshalJSON custom unmarshaller for Notify
func (n *Notify) UnmarshalJSON(data []byte) error {
	// Define a struct that mirrors Notify but with Integration as json.RawMessage
	// to defer its unmarshalling
	type Alias Notify
	temp := &struct {
		Integration json.RawMessage `json:"integration"`
		*Alias
	}{
		Alias: (*Alias)(n), // Cast n to Alias type to unmarshal other fields
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Now, determine the type of Integration and unmarshal accordingly
	// This assumes your JSON contains some type identifier for the integration.
	// You might need a temporary struct to peek into the raw JSON to read that identifier
	var typeIndicator struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(temp.Integration, &typeIndicator); err != nil {
		return err
	}

	switch typeIndicator.Type {
	case "slack":
		var slackIntegration SlackIntegration
		if err := json.Unmarshal(temp.Integration, &slackIntegration); err != nil {
			return err
		}
		n.Integration = &slackIntegration
	case "email":
		var emailIntegration EmailIntegration
		if err := json.Unmarshal(temp.Integration, &emailIntegration); err != nil {
			return err
		}
		n.Integration = &emailIntegration
	case "webform":
		var webformIntegration WebformIntegration
		if err := json.Unmarshal(temp.Integration, &webformIntegration); err != nil {
			return err
		}
		n.Integration = &webformIntegration
	default:
		return perr.InternalWithMessage(fmt.Sprintf("unknown integration type: %s", typeIndicator.Type))
	}

	return nil
}

func (n *Notify) MapInterface() (map[string]interface{}, error) {
	notifyMap := make(map[string]interface{})

	if n.Cc != nil {
		notifyMap["cc"] = n.Cc
	}

	if n.Bcc != nil {
		notifyMap["bcc"] = n.Bcc
	}

	if n.Channel != nil {
		notifyMap["channel"] = *n.Channel
	}

	if n.Description != nil {
		notifyMap["description"] = *n.Description
	}

	if n.Subject != nil {
		notifyMap["subject"] = *n.Subject
	}

	if n.Title != nil {
		notifyMap["title"] = *n.Title
	}

	if n.To != nil {
		notifyMap["to"] = n.To
	}

	var err error
	notifyMap["integration"], err = n.Integration.MapInterface()
	if err != nil {
		return nil, err
	}

	return notifyMap, nil
}
func (n *Notify) CtyValue() (cty.Value, error) {
	notifyMap := make(map[string]interface{})

	var err error

	if n.Cc != nil {
		notifyMap["cc"] = n.Cc
	}

	if n.Bcc != nil {
		notifyMap["bcc"] = n.Bcc
	}

	if n.Channel != nil {
		notifyMap["channel"] = *n.Channel
	}

	if n.Description != nil {
		notifyMap["description"] = *n.Description
	}

	if n.Subject != nil {
		notifyMap["subject"] = *n.Subject
	}

	if n.Title != nil {
		notifyMap["title"] = *n.Title
	}

	if n.To != nil {
		notifyMap["to"] = n.To
	}

	notifyMap["integration"], err = n.Integration.MapInterface()
	if err != nil {
		return cty.NilVal, err
	}

	ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(notifyMap)
	if err != nil {
		return cty.NilVal, err
	}

	return ctyVal, nil
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
