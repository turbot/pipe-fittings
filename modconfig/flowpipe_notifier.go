package modconfig

import "github.com/zclconf/go-cty/cty"

type Notifier struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Notifies []Notify
}

type Notify struct {
	Integration cty.Value `json:"integration" cty:"integration" hcl:"integration"`

	Cc          []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc         []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel     *string  `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Description *string  `json:"description,omitempty" cty:"description" hcl:"description,optional"`
	Subject     *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	Title       *string  `json:"title,omitempty" cty:"title" hcl:"title,optional"`
	To          *string  `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}
