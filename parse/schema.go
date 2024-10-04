package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/schema"
)

// cache resource schemas
var resourceSchemaCache = make(map[string]*hcl.BodySchema)

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
		{
			Type:       schema.BlockTypePartition,
			LabelNames: []string{schema.LabelType, schema.LabelName},
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

// DashboardBlockSchema is only used to validate the blocks of a Dashboard
var DashboardBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeInput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
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
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
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
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       "with",
			LabelNames: []string{schema.LabelName},
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
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       "category",
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       "with",
			LabelNames: []string{schema.LabelName},
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
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "validation",
		},
	},
}
