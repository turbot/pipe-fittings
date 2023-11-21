package schema

import "github.com/turbot/go-kit/helpers"

// NOTE: when adding a block type, be sure to update  QueryProviderBlocks/ReferenceBlocks/AllBlockTypes as needed
const (
	// require blocks
	BlockTypeSteampipe = "steampipe"
	BlockTypeMod       = "mod"
	BlockTypePlugin    = "plugin"
	// resource blocks
	BlockTypeQuery          = "query"
	BlockTypeControl        = "control"
	BlockTypeBenchmark      = "benchmark"
	BlockTypeDashboard      = "dashboard"
	BlockTypeContainer      = "container"
	BlockTypeChart          = "chart"
	BlockTypeCard           = "card"
	BlockTypeFlow           = "flow"
	BlockTypeGraph          = "graph"
	BlockTypeHierarchy      = "hierarchy"
	BlockTypeImage          = "image"
	BlockTypeInput          = "input"
	BlockTypeTable          = "table"
	BlockTypeText           = "text"
	BlockTypeLocals         = "locals"
	BlockTypeVariable       = "variable"
	BlockTypeParam          = "param"
	BlockTypeRequire        = "require"
	BlockTypeNode           = "node"
	BlockTypeEdge           = "edge"
	BlockTypeLegacyRequires = "requires"
	BlockTypeCategory       = "category"
	BlockTypeWith           = "with"
	BlockTypeError          = "error"

	// config blocks
	BlockTypeRateLimiter       = "limiter"
	BlockTypeConnection        = "connection"
	BlockTypeOptions           = "options"
	BlockTypeWorkspaceProfile  = "workspace"
	BlockTypePipeline          = "pipeline"
	BlockTypePipelineStep      = "step"
	BlockTypePipelineOutput    = "output"
	BlockTypeTrigger           = "trigger"
	BlockTypePipelineBasicAuth = "basic_auth"
	BlockTypeIntegration       = "integration"
	BlockTypeNotify            = "notify"
	BlockTypeLoop              = "loop"
	BlockTypeRetry             = "retry"
	BlockTypeThrow             = "throw"

	AttributeTypeValue     = "value"
	AttributeTypeSensitive = "sensitive"

	AttributeTypeType    = "type"
	AttributeTypeDefault = "default"

	// Pipeline param block
	AttributeTypeOptional = "optional"

	// Pipeline blocks
	BlockTypePipelineStepHttp      = "http"
	BlockTypePipelineStepSleep     = "sleep"
	BlockTypePipelineStepEmail     = "email"
	BlockTypePipelineStepEcho      = "echo"
	BlockTypePipelineStepTransform = "transform"
	BlockTypePipelineStepQuery     = "query"
	BlockTypePipelineStepExec      = "exec"
	BlockTypePipelineStepPipeline  = "pipeline"
	BlockTypePipelineStepFunction  = "function"
	BlockTypePipelineStepContainer = "container"
	BlockTypePipelineStepInput     = "input"

	// error block
	AttributeTypeIgnore  = "ignore"
	AttributeTypeRetries = "retries"

	// Common step attributes
	AttributeTypeTitle       = "title"
	AttributeTypeDependsOn   = "depends_on"
	AttributeTypeForEach     = "for_each"
	AttributeTypeDescription = "description"
	AttributeTypeIf          = "if"
	AttributeTypeUsername    = "username"
	AttributeTypePassword    = "password"
	AttributeTypeStepName    = "step_name"

	// pipeline attributes
	AttributeTypeTags            = "tags"
	AttributeTypeDocumentation   = "documentation"
	AttributeTypeUnqualifiedName = "unqualified_name"
	AttributeTypeName            = "name"
	AttributeTypeShortName       = "short_name"

	AttributeTypeStartedAt  = "started_at"
	AttributeTypeFinishedAt = "finished_at"

	// Used by query step
	AttributeTypeSql              = "sql"
	AttributeTypeArgs             = "args"
	AttributeTypeQuery            = "query"
	AttributeTypeRows             = "rows"
	AttributeTypeConnectionString = "connection_string"

	// Used by email step
	AttributeTypeBcc              = "bcc"
	AttributeTypeBody             = "body"
	AttributeTypeCc               = "cc"
	AttributeTypeContentType      = "content_type"
	AttributeTypeFrom             = "from"
	AttributeTypeHost             = "host"
	AttributeTypePort             = "port"
	AttributeTypeSenderCredential = "sender_credential" //nolint:gosec // Getting Potential hardcoded credentials warning
	AttributeTypeSenderName       = "sender_name"
	AttributeTypeSubject          = "subject"
	AttributeTypeTo               = "to"

	AttributeTypeToken         = "token"
	AttributeTypeSigningSecret = "signing_secret"
	AttributeTypeWebhookUrl    = "webhook_url"
	AttributeTypeNotifies      = "notifies"

	AttributeTypeIntegration = "integration"
	AttributeTypeChannel     = "channel"

	// Used by sleep step
	AttributeTypeDuration = "duration"

	// Used by http step
	AttributeTypeUrl              = "url"
	AttributeTypeMethod           = "method"
	AttributeTypeRequestBody      = "request_body"
	AttributeTypeRequestHeaders   = "request_headers"
	AttributeTypeRequestTimeoutMs = "request_timeout_ms"
	AttributeTypeCaCertPem        = "ca_cert_pem"
	AttributeTypeInsecure         = "insecure"
	AttributeTypeResponseHeaders  = "response_headers"
	AttributeTypeResponseBody     = "response_body"
	AttributeTypeStatusCode       = "status_code"
	AttributeTypeStatus           = "status"

	// Used by echo step
	AttributeTypeText    = "text"
	AttributeTypeNumeric = "numeric"
	AttributeTypeJson    = "json"

	// Used byy Pipeline step
	AttributeTypePipeline = "pipeline"

	// Used by input step
	AttributeTypeDefaultRecipient = "default_recipient"
	AttributeTypeDefaultSubject   = "default_subject"
	AttributeTypeOptions          = "options"
	AttributeTypeResponseUrl      = "response_url"
	AttributeTypeSmtpHost         = "smtp_host"
	AttributeTypeSmtpPassword     = "smtp_password"
	AttributeTypeSmtpPort         = "smtp_port"
	AttributeTypeSmtpServer       = "smtp_server"
	AttributeTypeSmtpTls          = "smtp_tls"
	AttributeTypeSmtpUsername     = "smtp_username"
	AttributeTypeSmtpsPort        = "smtps_port"

	AttributeTypeMessage = "message"

	// Functions attributes
	AttributeTypeRuntime  = "runtime"
	AttributeTypeEnv      = "env"
	AttributeTypeSrc      = "src"
	AttributeTypeHandler  = "handler"
	AttributeTypeFunction = "function"
	AttributeTypeEvent    = "event"

	AttributeTypeImage             = "image"
	AttributeTypeSource            = "source"
	AttributeTypeUser              = "user"
	AttributeTypeWorkdir           = "workdir"
	AttributeTypeCmd               = "cmd"
	AttributeTypeEntryPoint        = "entrypoint"
	AttributeTypeTimeout           = "timeout"
	AttributeTypeMemory            = "memory"
	AttributeTypeMemoryReservation = "memory_reservation"
	AttributeTypeMemorySwap        = "memory_swap"
	AttributeTypeMemorySwappiness  = "memory_swappiness"
	AttributeTypeReadOnly          = "read_only"

	// Trigger attributes
	AttributeTypeSchedule   = "schedule"
	AttributeTypePrimaryKey = "primary_key"
	AttributeTypeEvents     = "events"

	// Input step attributes
	AttributeTypePrompt    = "prompt"
	AttributeTypeSlackType = "slack_type"

	// All Possible Trigger Types
	TriggerTypeSchedule = "schedule"
	TriggerTypeInterval = "interval"
	TriggerTypeQuery    = "query"
	TriggerTypeHttp     = "http"

	// Integration Types
	IntegrationTypeSlack = "slack"
	IntegrationTypeEmail = "email"
	IntegrationTypeTeams = "teams"

	LabelName = "name"
	LabelType = "type"

	ResourceTypeSnapshot = "snapshot"
	AttributeArgs        = "args"
	AttributeQuery       = "query"

	// TODO KAI should these be AttributeType <MISC>
	AttributeVar   = "var"
	AttributeLocal = "local"

	AttributeEach = "each"
	AttributeKey  = "key"
)

// QueryProviderBlocks is a list of block types which implement QueryProvider
var QueryProviderBlocks = []string{
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeControl,
	BlockTypeEdge,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeQuery,
	BlockTypeNode,
	BlockTypeTable,
	BlockTypeWith,
}

// NodeAndEdgeProviderBlocks is a list of block types which implementnodeAndEdgeProvider
var NodeAndEdgeProviderBlocks = []string{
	BlockTypeHierarchy,
	BlockTypeFlow,
	BlockTypeGraph,
}

// ReferenceBlocks is a list of block types we store references for
var ReferenceBlocks = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeParam,
	BlockTypeCategory,
	BlockTypeWith,
}

var ValidResourceItemTypes = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeChart,
	BlockTypeCard,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeLocals,
	BlockTypeVariable,
	BlockTypeParam,
	BlockTypeRequire,
	BlockTypeNode,
	BlockTypeEdge,
	BlockTypeLegacyRequires,
	BlockTypeCategory,
	BlockTypeConnection,
	BlockTypeOptions,
	BlockTypeWorkspaceProfile,
	BlockTypePipeline,
	BlockTypeTrigger,
	BlockTypeWith,
	BlockTypeIntegration,
	// local is not an actual block name but is a resource type
	"local",
	// references
	"ref",
	// variables
	"var",
}

func IsValidResourceItemType(blockType string) bool {
	return helpers.StringSliceContains(ValidResourceItemTypes, blockType)
}
