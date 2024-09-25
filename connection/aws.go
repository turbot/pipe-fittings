package connection

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const (
	AwsConnectionType = "aws"
	defaultAwsTtl     = 5 * 60
)

type AwsConnection struct {
	ConnectionImpl

	AccessKey    *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey    *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	SessionToken *string `json:"session_token,omitempty" cty:"session_token" hcl:"session_token,optional"`
	Profile      *string `json:"profile,omitempty" cty:"profile" hcl:"profile,optional"`
}

func NewAwsConnection(shortName string, declRange hcl.Range) PipelingConnection {
	res := &AwsConnection{
		ConnectionImpl: NewConnectionImpl(AwsConnectionType, shortName, declRange),
	}
	// set the default ttl
	res.SetTtl(defaultAwsTtl)
	return res
}

func (c *AwsConnection) GetConnectionType() string {
	return AwsConnectionType
}

func (c *AwsConnection) Resolve(ctx context.Context) (PipelingConnection, error) {
	// if pipes metadata is set, call pipes to retrieve the creds
	if c.Pipes != nil {
		return c.Pipes.Resolve(ctx, &AwsConnection{})
	}

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
	newConnection := &AwsConnection{
		ConnectionImpl: c.ConnectionImpl,
		AccessKey:      &creds.AccessKeyID,
		SecretKey:      &creds.SecretAccessKey,
	}

	if creds.SessionToken != "" {
		newConnection.SessionToken = &creds.SessionToken
	}

	return newConnection, nil
}

func (c *AwsConnection) Equals(otherConnection PipelingConnection) bool {
	// If both pointers are nil, they are considered equal
	if c == nil && helpers.IsNil(otherConnection) {
		return true
	}

	if (c == nil && !helpers.IsNil(otherConnection)) || (c != nil && helpers.IsNil(otherConnection)) {
		return false
	}

	other, ok := otherConnection.(*AwsConnection)
	if !ok {
		return false
	}

	if !utils.PtrEqual(c.AccessKey, other.AccessKey) {
		return false
	}

	if !utils.PtrEqual(c.SecretKey, other.SecretKey) {
		return false
	}

	if !utils.PtrEqual(c.Profile, other.Profile) {
		return false
	}

	return c.GetConnectionImpl().Equals(otherConnection.GetConnectionImpl())
}

func (c *AwsConnection) Validate() hcl.Diagnostics {
	if c.Pipes != nil && (c.AccessKey != nil || c.SecretKey != nil || c.Profile != nil || c.SessionToken != nil) {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "if pipes block is defined, no other auth properties should be set",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	if c.AccessKey != nil && c.SecretKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "access_key defined without secret_key",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	if c.SecretKey != nil && c.AccessKey == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "secret_key defined without access_key",
				Subject:  c.DeclRange.HclRangePointer(),
			},
		}
	}

	return hcl.Diagnostics{}
}

func (c *AwsConnection) CtyValue() (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *AwsConnection) GetEnv() map[string]cty.Value {
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
