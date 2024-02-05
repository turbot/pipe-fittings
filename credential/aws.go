package credential

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type AwsCredential struct {
	CredentialImpl

	AccessKey    *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey    *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	SessionToken *string `json:"session_token,omitempty" cty:"session_token" hcl:"session_token,optional"`
	Ttl          *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
	Profile      *string `json:"profile,omitempty" cty:"profile" hcl:"profile,optional"`
}

func (c *AwsCredential) Validate() hcl.Diagnostics {

	if c.AccessKey != nil && c.SecretKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "access_key defined without secret_key",
				Subject:  &c.DeclRange,
			},
		}
	}

	if c.SecretKey != nil && c.AccessKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "secret_key defined without access_key",
				Subject:  &c.DeclRange,
			},
		}
	}

	return hcl.Diagnostics{}
}

func (c *AwsCredential) Resolve(ctx context.Context) (Credential, error) {

	// if access key and secret key are provided, just return it
	if c.AccessKey != nil && c.SecretKey != nil {
		return c, nil
	}

	var cfg aws.Config
	var err error

	// Load the AWS configuration from the shared credentials file
	if c.Profile != nil {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(*c.Profile), config.WithRegion("us-east-1"))
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
		if err != nil {
			return nil, err
		}

	}

	// Access the credentials from the configuration
	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return nil, err
	}

	// Don't modify existing credential, resolve to a new one
	newCreds := &AwsCredential{
		CredentialImpl: c.CredentialImpl,
		Ttl:            c.Ttl,
		AccessKey:      &creds.AccessKeyID,
		SecretKey:      &creds.SecretAccessKey,
	}

	if creds.SessionToken != "" {
		newCreds.SessionToken = &creds.SessionToken
	}

	return newCreds, nil
}

// in seconds
func (c *AwsCredential) GetTtl() int {
	if c.Ttl == nil {
		return 5 * 60
	}
	return *c.Ttl
}

func (c *AwsCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.AccessKey != nil {
		env["AWS_ACCESS_KEY_ID"] = cty.StringVal(*c.AccessKey)
	}
	if c.SecretKey != nil {
		env["AWS_SECRET_ACCESS_KEY"] = cty.StringVal(*c.SecretKey)
	}
	if c.SessionToken != nil {
		env["AWS_SESSION_TOKEN"] = cty.StringVal(*c.SessionToken)
	}
	return env
}

func (c *AwsCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := modconfig.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}
