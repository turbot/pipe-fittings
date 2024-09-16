package app_specific_connection

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
)

type ConnectionFunc func(block *hcl.Block) connection.PipelingConnection

var ConnectionTypeRegistry map[string]ConnectionFunc
