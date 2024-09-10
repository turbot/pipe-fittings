package workspace_profile

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
)

type WorkspaceProfile interface {
	SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics
	Name() string
	ShortName() string
	CtyValue() (cty.Value, error)
	OnDecoded() hcl.Diagnostics
	ConfigMap(cmd *cobra.Command) map[string]interface{}
	GetDeclRange() *hcl.Range

	GetOptionsForBlock(*hcl.Block) (options.Options, hcl.Diagnostics)

	GetInstallDir() *string
	// IsNil implements slightly hacky way of doing nil checks with generic types
	IsNil() bool
}

func NewWorkspaceProfile[T WorkspaceProfile](block *hcl.Block) (T, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var empty T
	var res any

	profileName := block.Labels[0]
	declRange := hclhelpers.BlockRange(block)
	switch any(empty).(type) {
	case *SteampipeWorkspaceProfile:
		res = &SteampipeWorkspaceProfile{ProfileName: profileName, DeclRange: declRange}
	case *FlowpipeWorkspaceProfile:
		res = &FlowpipeWorkspaceProfile{ProfileName: profileName, DeclRange: declRange}
	case *PowerpipeWorkspaceProfile:
		res = &PowerpipeWorkspaceProfile{ProfileName: profileName, DeclRange: declRange}
	case *TpWorkspaceProfile:
		res = &TpWorkspaceProfile{ProfileName: profileName, DeclRange: declRange}
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unsupported WorkspaceProfile type '%s'", reflect.TypeOf(empty).Name()),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return empty, diags
	}

	return res.(T), nil
}

func NewDefaultWorkspaceProfile[T WorkspaceProfile]() (T, hcl.Diagnostics) {
	return NewWorkspaceProfile[T](&hcl.Block{
		Type:      "workspace",
		Labels:    []string{"default"},
		DefRange:  hcl.Range{},
		TypeRange: hcl.Range{},
	})
}
