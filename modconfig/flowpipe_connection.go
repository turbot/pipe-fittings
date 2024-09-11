package modconfig

import (
	"context"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

var connectionTypRegistry = map[string]reflect.Type{
	(&AbuseIPDBConnection{}).GetConnectionType(): reflect.TypeOf(AbuseIPDBConnection{}),
	(&AlicloudConnection{}).GetConnectionType():  reflect.TypeOf(AlicloudConnection{}),
	(&AwsConnection{}).GetConnectionType():       reflect.TypeOf(AwsConnection{}),
	"azure":                                      reflect.TypeOf(AzureConnection{}),
	"bitbucket":                                  reflect.TypeOf(BitbucketConnection{}),
	"clickup":                                    reflect.TypeOf(ClickUpConnection{}),
	"datadog":                                    reflect.TypeOf(DatadogConnection{}),
	"discord":                                    reflect.TypeOf(DiscordConnection{}),
	"freshdesk":                                  reflect.TypeOf(FreshdeskConnection{}),
	"gcp":                                        reflect.TypeOf(GcpConnection{}),
	"github":                                     reflect.TypeOf(GithubConnection{}),
	"gitlab":                                     reflect.TypeOf(GitLabConnection{}),
	"ip2locationio":                              reflect.TypeOf(IP2LocationIOConnection{}),
	"ipstack":                                    reflect.TypeOf(IPstackConnection{}),
	"jira":                                       reflect.TypeOf(JiraConnection{}),
	"jumpcloud":                                  reflect.TypeOf(JumpCloudConnection{}),
	"mastodon":                                   reflect.TypeOf(MastodonConnection{}),
	"microsoft_teams":                            reflect.TypeOf(MicrosoftTeamsConnection{}),
	"okta":                                       reflect.TypeOf(OktaConnection{}),
	"openai":                                     reflect.TypeOf(OpenAIConnection{}),
	"opsgenie":                                   reflect.TypeOf(OpsgenieConnection{}),
	"pagerduty":                                  reflect.TypeOf(PagerDutyConnection{}),
	"sendgrid":                                   reflect.TypeOf(SendGridConnection{}),
	"servicenow":                                 reflect.TypeOf(ServiceNowConnection{}),
	"slack":                                      reflect.TypeOf(SlackConnection{}),
	"trello":                                     reflect.TypeOf(TrelloConnection{}),
	"turbot_guardrails":                          reflect.TypeOf(GuardrailsConnection{}),
	"turbot_pipes":                               reflect.TypeOf(PipesConnection{}),
	"uptime_robot":                               reflect.TypeOf(UptimeRobotConnection{}),
	"urlscan":                                    reflect.TypeOf(UrlscanConnection{}),
	"vault":                                      reflect.TypeOf(VaultConnection{}),
	"virus_total":                                reflect.TypeOf(VirusTotalConnection{}),
	"zendesk":                                    reflect.TypeOf(ZendeskConnection{}),
}

type PipelingConnection interface {
	HclResource
	ResourceWithMetadata

	SetHclResourceImpl(hclResourceImpl HclResourceImpl)
	GetConnectionType() string
	GetUnqualifiedName() string

	CtyValue() (cty.Value, error)
	Resolve(ctx context.Context) (PipelingConnection, error)
	GetTtl() int // in seconds

	Validate() hcl.Diagnostics
	getEnv() map[string]cty.Value

	Equals(PipelingConnection) bool
}

func ConnectionCtyType(connectionType string) cty.Type {
	goType := connectionTypRegistry[connectionType]
	if goType == nil {
		return cty.NilType
	}

	return cty.Capsule(connectionType, goType)
}

func NewPipelingConnection(block *hcl.Block) (PipelingConnection, error) {
	connectionType := block.Labels[0]
	connectionName := block.Labels[1]

	hclResourceImpl := NewHclResourceImplNoMod(block, connectionType, connectionName)

	conn, err := instantiateConnection(connectionType, hclResourceImpl)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func instantiateConnection(key string, hclResourceImpl HclResourceImpl) (PipelingConnection, error) {
	t, exists := connectionTypRegistry[key]
	if !exists {
		return nil, perr.BadRequestWithMessage("Invalid connection type " + key)
	}
	credInterface := reflect.New(t).Interface()
	cred, ok := credInterface.(PipelingConnection)
	if !ok {
		return nil, perr.InternalWithMessage("Failed to create connection")
	}
	cred.SetHclResourceImpl(hclResourceImpl)

	return cred, nil
}

type ConnectionImpl struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (c *ConnectionImpl) GetUnqualifiedName() string {
	return c.HclResourceImpl.UnqualifiedName
}

func (c *ConnectionImpl) SetHclResourceImpl(hclResourceImpl HclResourceImpl) {
	c.HclResourceImpl = hclResourceImpl
}

func DefaultPipelingConnections() (map[string]PipelingConnection, error) {
	conns := make(map[string]PipelingConnection)

	for k := range connectionTypRegistry {
		hclResourceImpl := HclResourceImpl{
			FullName:        k + ".default",
			ShortName:       "default",
			UnqualifiedName: k + ".default",
		}

		defaultCred, err := instantiateConnection(k, hclResourceImpl)
		if err != nil {
			return nil, err
		}

		conns[k+".default"] = defaultCred

		error_helpers.RegisterConnectionType(k)
	}

	return conns, nil
}

func customTypeValidationSingle(attr *hcl.Attribute, ctyVal cty.Value, ctyType cty.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	var valueMap map[string]cty.Value
	if ctyVal.Type().IsMapType() || ctyVal.Type().IsObjectType() {
		valueMap = ctyVal.AsValueMap()
	}

	if valueMap == nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "value must be a map if the type is a capsule",
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}

		return append(diags, diag)
	}

	encapsulatedType := ctyType.EncapsulatedType()
	encapulatedInstanceNew := reflect.New(encapsulatedType)
	if connInterface, ok := encapulatedInstanceNew.Interface().(PipelingConnection); ok {
		if valueMap["type"] == cty.NilVal {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "missing type in value",
			}
			if attr != nil {
				diag.Subject = &attr.Range
			}
			return append(diags, diag)
		}

		if connInterface.GetConnectionType() == valueMap["type"].AsString() {
			return diags
		}
	} else if ctyType.EncapsulatedType().String() == "*modconfig.ConnectionImpl" {
		if valueMap["resource_type"] == cty.NilVal {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "missing resource_type in value",
			}
			if attr != nil {
				diag.Subject = &attr.Range
			}
			return append(diags, diag)
		}

		if valueMap["resource_type"].AsString() == schema.BlockTypeConnection {
			return diags
		}
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "value type mismatched with the capsule type",
	}
	if attr != nil {
		diag.Subject = &attr.Range
	}

	return append(diags, diag)
}

func customTypeValidationList(attr *hcl.Attribute, ctyVal cty.Value, ctyType cty.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	if !ctyVal.Type().IsTupleType() && !ctyVal.Type().IsCollectionType() && !ctyVal.Type().IsListType() {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "default value must be a list if the type is a list of capsules",
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}

		return append(diags, diag)
	}

	encapsulatedType := ctyType.ListElementType().EncapsulatedType()
	encapulatedInstanceNew := reflect.New(encapsulatedType)
	if connInterface, ok := encapulatedInstanceNew.Interface().(PipelingConnection); ok {
		for _, val := range ctyVal.AsValueSlice() {
			if val.Type().IsMapType() || val.Type().IsObjectType() {
				valueMap := val.AsValueMap()
				if valueMap["type"] == cty.NilVal {
					diag := &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "missing type in default value",
					}
					if attr != nil {
						diag.Subject = &attr.Range
					}
					return append(diags, diag)
				}

				if connInterface.GetConnectionType() != valueMap["type"].AsString() {
					diag := &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "default value type mismatched with the capsule type",
					}
					if attr != nil {
						diag.Subject = &attr.Range
					}

					return append(diags, diag)

				}
			} else {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "default value must be a map if the type is a list of capsules",
				}
				if attr != nil {
					diag.Subject = &attr.Range
				}

				return append(diags, diag)
			}
		}
	} else if ctyType.ListElementType().EncapsulatedType().String() == "*modconfig.ConnectionImpl" {
		for _, val := range ctyVal.AsValueSlice() {
			if val.Type().IsMapType() || val.Type().IsObjectType() {
				valueMap := val.AsValueMap()

				if valueMap["resource_type"] == cty.NilVal {
					diag := &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "missing resource_type in value",
					}
					if attr != nil {
						diag.Subject = &attr.Range
					}
					return append(diags, diag)
				}

				if valueMap["resource_type"].AsString() != schema.BlockTypeConnection {
					diag := &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "value type mismatched with the capsule type",
					}
					if attr != nil {
						diag.Subject = &attr.Range
					}
					return append(diags, diag)
				}
			} else {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "value must be a map if the type is a list of capsules",
				}
				if attr != nil {
					diag.Subject = &attr.Range
				}
				return append(diags, diag)
			}
		}
	}

	return diags
}

func CustomTypeValidation(attr *hcl.Attribute, ctyVal cty.Value, ctyType cty.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// It must be a capsule type OR a list where the element type is a capsule
	if !ctyType.IsCapsuleType() && !(ctyType.IsListType() && ctyType.ListElementType().IsCapsuleType()) {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Type must be a capsule",
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}
		diags = append(diags, diag)
		return diags
	}

	if ctyType.IsCapsuleType() {
		return customTypeValidationSingle(attr, ctyVal, ctyType)
	}

	// must be a list then
	return customTypeValidationList(attr, ctyVal, ctyType)
}
