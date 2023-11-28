package modconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Credential interface {
	HclResource
	ResourceWithMetadata

	GetCredentialType() string
	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (Credential, error)

	Validate() hcl.Diagnostics
	getEnv() map[string]cty.Value
}

type AwsCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	AccessKey    *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey    *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	SessionToken *string `json:"session_token,omitempty" cty:"session_token" hcl:"session_token,optional"`
	Ttl          *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
	Profile      *string `json:"profile,omitempty" cty:"profile" hcl:"profile,optional"`
}

func DefaultCredentialNames() []string {
	return []string{"aws.default", "slack.default", "basic.default", "gcp.default", "aws.<dynamic>", "slack.<dynamic", "basic.<dynamic>", "gcp.<dynamic>"}
}

func (*AwsCredential) GetCredentialType() string {
	return "aws"
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

	newCreds := &AwsCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        c.FullName,
			UnqualifiedName: c.UnqualifiedName,
			DeclRange:       c.DeclRange,
			blockType:       c.blockType,
		},
		Type:      c.Type,
		Ttl:       c.Ttl,
		AccessKey: &creds.AccessKeyID,
		SecretKey: &creds.SecretAccessKey,
	}

	if creds.SessionToken != "" {
		newCreds.SessionToken = &creds.SessionToken
	}

	return newCreds, nil
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
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

type SlackCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Token *string `json:"token,omitempty" cty:"token" hcl:"token,optional"`
}

func (*SlackCredential) GetCredentialType() string {
	return "slack"
}

func (c *SlackCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	if c.Token != nil {
		env["SLACK_TOKEN"] = cty.StringVal(*c.Token)
	}
	return env
}

func (c *SlackCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *SlackCredential) Resolve(ctx context.Context) (Credential, error) {
	return c, nil
}

func (c *SlackCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type GcpCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Credentials *string `json:"credentials,omitempty" cty:"credentials" hcl:"credentials,optional"`
	Ttl         *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
}

func (*GcpCredential) GetCredentialType() string {
	return "gcp"
}

func (c *GcpCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}

func (c *GcpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *GcpCredential) Resolve(ctx context.Context) (Credential, error) {
	return c, nil
}

func (c *GcpCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

type BasicCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (c *BasicCredential) getEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}

func (*BasicCredential) GetCredentialType() string {
	return "basic"
}

func (c *BasicCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.getEnv())

	return cty.ObjectVal(valueMap), nil
}

func (c *BasicCredential) Resolve(ctx context.Context) (Credential, error) {
	return c, nil
}

func (c *BasicCredential) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func DefaultCredentials() map[string]Credential {
	credentials := make(map[string]Credential)

	credentials["aws.default"] = &AwsCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "aws.default",
			ShortName:       "default",
			UnqualifiedName: "aws.default",
		},
		Type: "aws",
	}
	credentials["slack.default"] = &SlackCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "slack.default",
			ShortName:       "default",
			UnqualifiedName: "slack.default",
		},
		Type: "slack",
	}
	credentials["gcp.default"] = &GcpCredential{
		HclResourceImpl: HclResourceImpl{
			FullName:        "gcp.default",
			ShortName:       "default",
			UnqualifiedName: "gcp.default",
		},
		Type: "gcp",
	}

	return credentials
}

func NewCredential(block *hcl.Block) Credential {

	credentialType := block.Labels[0]
	credentialName := block.Labels[1]

	credentialFullName := credentialType + "." + credentialName

	if credentialType == "aws" {
		credential := &AwsCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "aws",
		}
		return credential
	} else if credentialType == "basic" {
		credential := &BasicCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "basic",
		}
		return credential
	} else if credentialType == "slack" {
		credential := &SlackCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialFullName,
				ShortName:       credentialName,
				UnqualifiedName: credentialFullName,
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "slack",
		}
		return credential
	}

	return nil
}
