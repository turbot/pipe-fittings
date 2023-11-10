package modconfig

type RetryConfig struct {
	If      *bool `json:"if,omitempty" hcl:"if,optional" cty:"if"`
	Retries int   `json:"retries" hcl:"retries" cty:"retries"`
}
