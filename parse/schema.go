package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/schema"
)

// cache resource schemas
var resourceSchemaCache = make(map[string]*hcl.BodySchema)

// Custom config schema for Flowpipe. The connection block setup is different, Steampipe only has one label
// while Flowipe has 2 labels. Credential, CredentialImport, Integration and Notifer are also specific to Flowpipe
var FlowpipeConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			// Flowpipe connnections have 2 labels
			Type:       schema.BlockTypeConnection,
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       schema.BlockTypeOptions,
			LabelNames: []string{"type"},
		},
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{"name"},
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
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{"name"},
		},
	},
}

var ConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeConnection,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypePlugin,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeOptions,
			LabelNames: []string{"type"},
		},
		{
			Type:       schema.BlockTypeWorkspaceProfile,
			LabelNames: []string{"name"},
		},
	},
}

var PluginBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeRateLimiter,
			LabelNames: []string{"name"},
		},
	},
}

var WorkspaceProfileBlockSchema = &hcl.BodySchema{

	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "options",
			LabelNames: []string{"type"},
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
			LabelNames: []string{"type"},
		},
	},
}

// WorkspaceBlockSchema is the top level schema for all workspace resources
var WorkspaceBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       string(schema.BlockTypeMod),
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeVariable,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeQuery,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeControl,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeBenchmark,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeDashboard,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeCard,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeChart,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeFlow,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeGraph,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeHierarchy,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeImage,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeInput,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeTable,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeText,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeNode,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeEdge,
			LabelNames: []string{"name"},
		},
		{
			Type: schema.BlockTypeLocals,
		},
		{
			Type:       schema.BlockTypeCategory,
			LabelNames: []string{"name"},
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

// DashboardBlockSchema is only used to validate the blocks of a Dashboard
var DashboardBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeInput,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{"name"},
		},
		{
			Type: schema.BlockTypeWith,
		},
		{
			Type: schema.BlockTypeContainer,
		},
		{
			Type: schema.BlockTypeCard,
		},
		{
			Type: schema.BlockTypeChart,
		},
		{
			Type: schema.BlockTypeBenchmark,
		},
		{
			Type: schema.BlockTypeControl,
		},
		{
			Type: schema.BlockTypeFlow,
		},
		{
			Type: schema.BlockTypeGraph,
		},
		{
			Type: schema.BlockTypeHierarchy,
		},
		{
			Type: schema.BlockTypeImage,
		},
		{
			Type: schema.BlockTypeTable,
		},
		{
			Type: schema.BlockTypeText,
		},
	},
}

// DashboardContainerBlockSchema is only used to validate the blocks of a DashboardContainer
var DashboardContainerBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeInput,
			LabelNames: []string{"name"},
		},
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{"name"},
		},
		{
			Type: schema.BlockTypeContainer,
		},
		{
			Type: schema.BlockTypeCard,
		},
		{
			Type: schema.BlockTypeChart,
		},
		{
			Type: schema.BlockTypeBenchmark,
		},
		{
			Type: schema.BlockTypeControl,
		},
		{
			Type: schema.BlockTypeFlow,
		},
		{
			Type: schema.BlockTypeGraph,
		},
		{
			Type: schema.BlockTypeHierarchy,
		},
		{
			Type: schema.BlockTypeImage,
		},
		{
			Type: schema.BlockTypeTable,
		},
		{
			Type: schema.BlockTypeText,
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

// QueryProviderBlockSchema schema for all blocks satisfying QueryProvider interface
// NOTE: these are just the blocks/attributes that are explicitly decoded
// other query provider properties are implicitly decoded using tags
var QueryProviderBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "param",
			LabelNames: []string{"name"},
		},
		{
			Type:       "with",
			LabelNames: []string{"name"},
		},
	},
}

// NodeAndEdgeProviderSchema is used to decode graph/hierarchy/flow
// (EXCEPT categories)
var NodeAndEdgeProviderSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "param",
			LabelNames: []string{"name"},
		},
		{
			Type:       "category",
			LabelNames: []string{"name"},
		},
		{
			Type:       "with",
			LabelNames: []string{"name"},
		},
		{
			Type: schema.BlockTypeNode,
		},
		{
			Type: schema.BlockTypeEdge,
		},
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
			Name: schema.AttributeTypeSubtype,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "validation",
		},
	},
}
