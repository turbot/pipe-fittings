package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/schema"
)

var FlowpipeConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
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
		{
			Type:       schema.BlockTypeCredential,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
	},
}

var IntegrationSlackBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeToken,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSigningSecret,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeWebhookUrl,
			Required: false,
		},
	},
}

var IntegrationEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSmtpTls,
			Required: false,
		},
		{
			Name: schema.AttributeTypeSmtpHost,
		},
		{
			Name: schema.AttributeTypeSmtpPort,
		},
		{
			Name:     schema.AttributeTypeSmtpsPort,
			Required: false,
		},
		{
			Name: schema.AttributeTypeSmtpUsername,
		},
		{
			Name: schema.AttributeTypeSmtpPassword,
		},
		{
			Name: schema.AttributeTypeFrom,
		},
		{
			Name:     schema.AttributeTypeDefaultRecipient,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDefaultSubject,
			Required: false,
		},
	},
}

var IntegrationTeamsBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
	},
}

var TriggerScheduleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSchedule,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePipeline,
			Required: true,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
}

var TriggerIntervalBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSchedule,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePipeline,
			Required: true,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
}

var TriggerQueryBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			// Schedule is not a required attribute for Query Trigger, default to every 15 minutes
			Name: schema.AttributeTypeSchedule,
		},
		{
			Name: schema.AttributeTypeSql,
		},
		{
			Name: schema.AttributeTypePrimaryKey,
		},
		{
			Name: schema.AttributeTypeConnectionString,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeCapture,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var TriggerHttpBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name:     schema.AttributeTypePipeline,
			Required: true,
		},
		{
			Name: schema.AttributeTypeExecutionMode,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
}

var PipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipelineStep,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var PipelineOutputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name:     schema.AttributeTypeValue,
			Required: true,
		},
	},
}

var PipelineParamBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeType,
		},
		{
			Name: schema.AttributeTypeDefault,
		},
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeOptional,
			Required: false,
		},
	},
}

var PipelineStepHttpBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeUrl,
			Required: true,
		},
		{
			Name: schema.AttributeTypeMethod,
		},
		{
			Name: schema.AttributeTypeCaCertPem,
		},
		{
			Name: schema.AttributeTypeInsecure,
		},
		{
			Name: schema.AttributeTypeRequestBody,
		},
		{
			Name: schema.AttributeTypeRequestHeaders,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypePipelineBasicAuth,
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineBasicAuthBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeUsername,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePassword,
			Required: true,
		},
	},
}

var PipelineStepSleepBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeDuration,
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeTo,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeFrom,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeSmtpUsername,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeSmtpPassword,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeHost,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePort,
			Required: true,
		},
		{
			Name: schema.AttributeTypeSenderName,
		},
		{
			Name: schema.AttributeTypeCc,
		},
		{
			Name: schema.AttributeTypeBcc,
		},
		{
			Name: schema.AttributeTypeBody,
		},
		{
			Name: schema.AttributeTypeContentType,
		},
		{
			Name: schema.AttributeTypeSubject,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepQueryBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeSql,
		},
		{
			Name: schema.AttributeTypeConnectionString,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepTransformBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeValue,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepPipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypePipeline,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepFunctionBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeSource,
			Required: true,
		},
		{
			Name: schema.AttributeTypeHandler,
		},
		{
			Name: schema.AttributeTypeRuntime,
		},
		{
			Name: schema.AttributeTypeEnv,
		},
		{
			Name: schema.AttributeTypeEvent,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepContainerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeImage,
			// Required: true,
		},
		{
			Name: schema.AttributeTypeSource,
			// Required: true,
		},
		{
			Name: schema.AttributeTypeCmd,
		},
		{
			Name: schema.AttributeTypeEnv,
		},
		{
			Name: schema.AttributeTypeEntryPoint,
		},
		{
			Name: schema.AttributeTypeCpuShares,
		},
		{
			Name: schema.AttributeTypeMemory,
		},
		{
			Name: schema.AttributeTypeMemoryReservation,
		},
		{
			Name: schema.AttributeTypeMemorySwap,
		},
		{
			Name: schema.AttributeTypeMemorySwappiness,
		},
		{
			Name: schema.AttributeTypeReadOnly,
		},
		{
			Name: schema.AttributeTypeUser,
		},
		{
			Name: schema.AttributeTypeWorkdir,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepInputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeOptions,
		},
		{
			Name: schema.AttributeTypePrompt,
		},
		{
			Name: schema.AttributeTypeNotifies,
		},
		{
			Name:     schema.AttributeTypeType,
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeNotify,
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepInputNotifyBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeIntegration,
		},
		{
			Name: schema.AttributeTypeChannel,
		},
	},
}
