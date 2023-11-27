package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Credential interface {
	HclResource
	ResourceWithMetadata

	GetCredentialType() string
	GetEnv() map[string]cty.Value
	CtyValue() (cty.Value, error)
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

func (c *AwsCredential) GetEnv() map[string]cty.Value {
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
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

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

func (c *SlackCredential) GetEnv() map[string]cty.Value {
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
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
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

func (c *GcpCredential) GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	return env
}

func (c *GcpCredential) CtyValue() (cty.Value, error) {
	ctyValue, err := GetCtyValue(c)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

type BasicCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (c *BasicCredential) GetEnv() map[string]cty.Value {
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
	valueMap["env"] = cty.ObjectVal(c.GetEnv())

	return cty.ObjectVal(valueMap), nil
}

func DefaultCredentials() map[string]Credential {
	credentials := make(map[string]Credential)
	// credentials["aws.default"] = &AwsCredential{
	// 	HclResourceImpl: HclResourceImpl{
	// 		FullName:        "aws.default",
	// 		UnqualifiedName: "credential.aws.default",
	// 	},
	// 	Type: "aws",
	// }
	// credentials["slack.default"] = &SlackCredential{
	// 	HclResourceImpl: HclResourceImpl{
	// 		FullName:        "slack.default",
	// 		UnqualifiedName: "credential.slack.default",
	// 	},
	// 	Type: "slack",
	// }
	// credentials["gcp.default"] = &GcpCredential{
	// 	HclResourceImpl: HclResourceImpl{
	// 		FullName:        "gcp.default",
	// 		UnqualifiedName: "credential.gcp.default",
	// 	},
	// 	Type: "gcp",
	// }

	return credentials
}

func NewCredential(block *hcl.Block) Credential {

	credentialType := block.Labels[0]
	credentialName := block.Labels[1]

	credentialName = credentialType + "." + credentialName

	if credentialType == "aws" {
		credential := &AwsCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialName,
				UnqualifiedName: "credential." + block.Labels[0] + "." + block.Labels[1],
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "aws",
		}
		return credential
	} else if credentialType == "basic" {
		credential := &BasicCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialName,
				UnqualifiedName: "credential." + block.Labels[0] + "." + block.Labels[1],
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "basic",
		}
		return credential
	} else if credentialType == "slack" {
		credential := &SlackCredential{
			HclResourceImpl: HclResourceImpl{
				FullName:        credentialName,
				UnqualifiedName: "credential." + block.Labels[0] + "." + block.Labels[1],
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "slack",
		}
		return credential
	}

	return nil
}
