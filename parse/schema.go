package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"reflect"
)

// cache resource schemas
var ResourceSchemaCache = make(map[string]*hcl.BodySchema)

// build the hcl schema for this resource
func getResourceSchema(resource modconfig.HclResource, nestedStructs []any) *hcl.BodySchema {
	t := reflect.TypeOf(helpers.DereferencePointer(resource))
	typeName := t.Name()

	if cachedSchema, ok := ResourceSchemaCache[typeName]; ok {
		return cachedSchema
	}
	var res = &hcl.BodySchema{}

	// ensure we cache before returning
	defer func() {
		ResourceSchemaCache[typeName] = res
	}()

	var schemas []*hcl.BodySchema

	// build schema for top level object
	schemas = append(schemas, GetSchemaForStruct(t))

	// now get schemas for any nested structs (using cache)
	for _, nestedStruct := range nestedStructs {
		t := reflect.TypeOf(helpers.DereferencePointer(nestedStruct))
		typeName := t.Name()

		// is this cached?
		nestedStructSchema, schemaCached := ResourceSchemaCache[typeName]
		if !schemaCached {
			nestedStructSchema = GetSchemaForStruct(t)
			ResourceSchemaCache[typeName] = nestedStructSchema
		}

		// add to our list of schemas
		schemas = append(schemas, nestedStructSchema)
	}

	// now merge the schemas
	for _, s := range schemas {
		res.Blocks = append(res.Blocks, s.Blocks...)
		res.Attributes = append(res.Attributes, s.Attributes...)
	}

	if resource.GetBlockType() == schema.BlockTypeMod {
		res.Blocks = append(res.Blocks, hcl.BlockHeaderSchema{Type: schema.BlockTypeRequire})
	}

	// if there is app specific schema, add it in
	if AppSpecificGetResourceSchemaFunc != nil {
		res = AppSpecificGetResourceSchemaFunc(resource, res)
	}
	return res
}

// PowerpipeConfigBlockSchema defines the config schema for Flowpipe config blocks.
// The connection block setup is different, Steampipe only has one label while Pipelingconnections has 2 labels.
// Credential, CredentialImport, Integration and Notifer are specific to Flowpipe
var PowerpipeConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			// Flowpipe connnections have 2 labels
			Type:       schema.BlockTypeConnection,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{"name"},
		},
	},
}

// FlowpipeConfigBlockSchema defines the config schema for Flowpipe config blocks.
// The connection block setup is different, Steampipe only has one label while Pipelingconnections has 2 labels.
// Credential, CredentialImport, Integration and Notifer are specific to Flowpipe
var FlowpipeConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			// Flowpipe connnections have 2 labels
			Type:       schema.BlockTypeConnection,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypeOptions,
			LabelNames: []string{"type"},
		},
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeCredential,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypeCredentialImport,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeConnectionImport,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeIntegration,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypeNotifier,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeOptions,
			LabelNames: []string{"type"},
		},
	},
}

var SteampipeConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeConnection,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypePlugin,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeOptions,
			LabelNames: []string{"type"},
		},
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var TpConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			// Tp connections have 2 labels
			Type:       schema.BlockTypeConnection,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypePartition,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypePlugin,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var PluginBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeRateLimiter,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var WorkspaceProfileBlockSchema = &hcl.BodySchema{

	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "options",
			LabelNames: []string{schema.LabelType},
		},
	},
}

var ConnectionBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "plugin",
			Required: true,
		},
		{
			Name: "type",
		},
		{
			Name: "connections",
		},
		{
			Name: "import_schema",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "options",
			LabelNames: []string{schema.LabelType},
		},
	},
}

// WorkspaceBlockSchema is the top level schema for all workspace resources
var WorkspaceBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       string(schema.BlockTypeMod),
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeVariable,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeQuery,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeControl,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeDetection,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeDetectionBenchmark,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeBenchmark,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeDashboard,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeCard,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeChart,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeFlow,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeGraph,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeHierarchy,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeImage,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeInput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeTable,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeText,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeNode,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeEdge,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLocals,
		},
		{
			Type:       schema.BlockTypeCategory,
			LabelNames: []string{schema.LabelName},
		},

		// Flowpipe
		{
			Type:       schema.BlockTypePipeline,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeTrigger,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypeIntegration,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
	},
}

var BenchmarkBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "children"},
		{Name: "description"},
		{Name: "documentation"},
		{Name: "tags"},
		{Name: "title"},
		// for report benchmark blocks
		{Name: "width"},
		{Name: "base"},
		{Name: "type"},
		{Name: "display"},
	},
}

var ParamDefBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "description"},
		{Name: "default"},
	},
}

var VariableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeDefault,
		},
		{
			Name: schema.AttributeTypeType,
		},
		{
			Name: schema.AttributeTypeSensitive,
		},
		{
			Name: schema.AttributeTypeTags,
		},
		{
			Name: schema.AttributeTypeEnum,
		},
		{
			Name: schema.AttributeTypeFormat,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "validation",
		},
	},
}
