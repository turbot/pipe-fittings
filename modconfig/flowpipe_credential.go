package modconfig

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type Credential interface {
	HclResource
	ResourceWithMetadata

	GetCredentialType() string
}

type AwsCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	AccessKey    *string `json:"access_key,omitempty" cty:"access_key" hcl:"access_key,optional"`
	SecretKey    *string `json:"secret_key,omitempty" cty:"secret_key" hcl:"secret_key,optional"`
	SessionToken *string `json:"session_token,omitempty" cty:"session_token" hcl:"session_token,optional"`
	Ttl          *int    `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
}

func (*AwsCredential) GetCredentialType() string {
	return "aws"
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

type BasicCredential struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Type string `json:"type" cty:"type" hcl:"type,label"`

	Username *string `json:"username,omitempty" cty:"username" hcl:"username,optional"`
	Password *string `json:"password,omitempty" cty:"password" hcl:"password,optional"`
}

func (*BasicCredential) GetCredentialType() string {
	return "basic"
}

func NewCredential(mod *Mod, block *hcl.Block) Credential {

	credentialType := block.Labels[0]
	credentialName := block.Labels[1]

	credentialName = credentialType + "." + credentialName

	if mod != nil {
		modName := mod.Name()
		if strings.HasPrefix(modName, "mod") {
			modName = strings.TrimPrefix(modName, "mod.")
		}
		credentialName = modName + ".credential." + credentialName
	} else {
		credentialName = "local.credential." + credentialName
	}

	if credentialType == "aws" {
		credential := &AwsCredential{
			HclResourceImpl: HclResourceImpl{
				// The FullName is the full name of the resource, including the mod name
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
				// The FullName is the full name of the resource, including the mod name
				FullName:        credentialName,
				UnqualifiedName: "credential." + block.Labels[0] + "." + block.Labels[1],
				DeclRange:       block.DefRange,
				blockType:       block.Type,
			},
			Type: "email",
		}
		return credential
	}

	return nil
}
