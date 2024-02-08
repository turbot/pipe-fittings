package modconfig

import "github.com/zclconf/go-cty/cty"

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

	integrationCty, err := defaultWebformIntegration.CtyValue()
	if err != nil {
		return nil, err
	}

	notify := Notify{
		Integration: integrationCty,
	}
	notifier.Notifies = []Notify{notify}

	notifiers["default"] = notifier

	return notifiers, nil
}

func (c *Notifier) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}
	return ctyValue, nil

	// valueMap := ctyValue.AsValueMap()
	// valueMap["env"] = cty.ObjectVal(c.getEnv())

	// return cty.ObjectVal(valueMap), nil
}

type Notify struct {
	Integration cty.Value `json:"integration" cty:"integration" hcl:"integration"`

	Cc          []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc         []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel     *string  `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Description *string  `json:"description,omitempty" cty:"description" hcl:"description,optional"`
	Subject     *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	Title       *string  `json:"title,omitempty" cty:"title" hcl:"title,optional"`
	To          []string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}
