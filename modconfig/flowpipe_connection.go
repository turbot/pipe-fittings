package modconfig

import (
	"context"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
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

func validateMapAttribute(attr *hcl.Attribute, valueMap map[string]cty.Value, key, errMsg string) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	if valueMap[key] == cty.NilVal {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  errMsg,
		}
		if attr != nil {
			diag.Subject = &attr.Range
		}
		return append(diags, diag)
	}
	return diags
}

func customTypeValidationSingle(attr *hcl.Attribute, ctyVal cty.Value, encapsulatedGoType reflect.Type) hcl.Diagnostics {
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

	encapulatedInstanceNew := reflect.New(encapsulatedGoType)
	if connInterface, ok := encapulatedInstanceNew.Interface().(PipelingConnection); ok {
		diags := validateMapAttribute(attr, valueMap, "type", "missing type in value")
		if len(diags) > 0 {
			return diags
		}

		if connInterface.GetConnectionType() == valueMap["type"].AsString() {
			return diags
		}
	} else if encapsulatedGoType.String() == "*modconfig.ConnectionImpl" {
		diags := validateMapAttribute(attr, valueMap, "resource_type", "missing resource_type in value")
		if len(diags) > 0 {
			return diags
		}

		if valueMap["resource_type"].AsString() == schema.BlockTypeConnection {
			return diags
		}
	} else if encapsulatedGoType.String() == "*modconfig.NotifierImpl" {
		diags := validateMapAttribute(attr, valueMap, "resource_type", "missing resource_type in value")
		if len(diags) > 0 {
			return diags
		}

		if valueMap["resource_type"].AsString() == schema.BlockTypeNotifier {
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

func customTypeCheckResourceTypeCorrect(attr *hcl.Attribute, val cty.Value, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	if val.Type().IsMapType() || val.Type().IsObjectType() {
		valueMap := val.AsValueMap()

		diags := validateMapAttribute(attr, valueMap, "resource_type", "missing resource_type in value")
		if len(diags) > 0 {
			return diags
		}

		encapsulatedInstanceNew := reflect.New(encapsulatedGoType)
		valid := false
		if pc, ok := encapsulatedInstanceNew.Interface().(PipelingConnection); ok {
			// Validate list of capsule type
			valid = valueMap["resource_type"].AsString() == schema.BlockTypeConnection && valueMap["type"].AsString() == pc.GetConnectionType()
		} else if encapsulatedGoType.String() == "*modconfig.ConnectionImpl" {
			valid = valueMap["resource_type"].AsString() == schema.BlockTypeConnection
		} else if encapsulatedGoType.String() == "*modconfig.NotifierImpl" {
			// Validate internal notifier resource
			valid = valueMap["resource_type"].AsString() == schema.BlockTypeNotifier
		}

		if !valid {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "value type mismatched with the capsule type",
			}
			if attr != nil {
				diag.Subject = &attr.Range
			}
			return append(diags, diag)
		}

		return diags
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "value must be a map if the type is a list of capsules",
	}
	if attr != nil {
		diag.Subject = &attr.Range
	}
	return append(diags, diag)

}

func customTypeValidation(attr *hcl.Attribute, ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// short circuit .. if it's object type we can't validate .. it's too complicated right now
	//
	// i.e. object(string, connection.aws, bool)
	if settingType.IsObjectType() {
		return diags
	}

	if ctyVal.Type().IsMapType() || ctyVal.Type().IsObjectType() {
		// Validate map or object type
		diags = customTypeValidateMapOrObject(attr, ctyVal, settingType, encapsulatedGoType)
	} else if hclhelpers.IsListLike(ctyVal.Type()) {
		// Validate list type, including nested lists or maps/objects
		diags = customTypeValidateList(attr, ctyVal, settingType, encapsulatedGoType)
	}

	return diags
}

func customTypeValidateMapOrObject(attr *hcl.Attribute, ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	valueMap := ctyVal.AsValueMap()

	for _, val := range valueMap {
		if val.Type().IsMapType() || val.Type().IsObjectType() {
			// Recursive validation for nested map/object types
			// does it have a resource type?
			innerValMap := val.AsValueMap()
			if innerValMap["resource_type"] != cty.NilVal {
				nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
				diags = append(diags, nestedDiags...)
				continue
			}

			nestedDiags := customTypeValidateMapOrObject(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else if hclhelpers.IsListLike(val.Type()) {
			// Recursive validation for nested list types
			nestedDiags := customTypeValidateList(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else {
			nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		}
	}

	return diags
}

func customTypeValidateList(attr *hcl.Attribute, ctyVal cty.Value, settingType cty.Type, encapsulatedGoType reflect.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for _, val := range ctyVal.AsValueSlice() {
		if hclhelpers.IsListLike(val.Type()) {
			// Recursive validation for nested list
			nestedDiags := customTypeValidateList(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else if val.Type().IsMapType() || val.Type().IsObjectType() {
			// Recursive validation for nested map/object inside list
			// does it have a resource type?
			innerValMap := val.AsValueMap()
			if innerValMap["resource_type"] != cty.NilVal {
				nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
				diags = append(diags, nestedDiags...)
				continue
			}

			nestedDiags := customTypeValidateMapOrObject(attr, val, settingType, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		} else {
			nestedDiags := customTypeCheckResourceTypeCorrect(attr, val, encapsulatedGoType)
			diags = append(diags, nestedDiags...)
		}
	}

	return diags
}

func CustomTypeValidation(attr *hcl.Attribute, ctyVal cty.Value, ctyType cty.Type) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// It must be a capsule type OR a list where the element type is a capsule
	encapsulatedGoType, ok := hclhelpers.IsNestedCapsuleType(ctyType)
	if !ok {
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
		return customTypeValidationSingle(attr, ctyVal, encapsulatedGoType)
	}

	return customTypeValidation(attr, ctyVal, ctyType, encapsulatedGoType)
}
