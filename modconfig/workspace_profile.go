package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/options"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

type WorkspaceProfile interface {
	SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics
	Name() string
	ShortName() string
	CtyValue() (cty.Value, error)
	OnDecoded() hcl.Diagnostics
	ConfigMap(cmd *cobra.Command) map[string]interface{}
	GetDeclRange() *hcl.Range

	// TODO do we actually need this in the interface or is it steampipe specific
	GetModLocation() *string
	GetInstallDir() *string
	// TODO slightly hacky way of doing nil checks with generic types
	IsNil() bool
}

func NewWorkspaceProfile[T WorkspaceProfile](block *hcl.Block) (T, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var empty T
	switch t := any(empty).(type) {
	case *SteampipeWorkspaceProfile:
		t = &SteampipeWorkspaceProfile{
			ProfileName: block.Labels[0],
			DeclRange:   hclhelpers.BlockRange(block),
		}
		return any(t).(T), nil
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unsupported WorkspaceProfile type '%s'", reflect.TypeOf(empty).Name()),
			Subject:  hclhelpers.BlockRangePointer(block),
		})
		return empty, diags
	}
}

func NewDefaultWorkspaceProfile[T WorkspaceProfile]() (T, hcl.Diagnostics) {
	return NewWorkspaceProfile[T](&hcl.Block{
		Type:      "workspace",
		Labels:    []string{"default"},
		DefRange:  hcl.Range{},
		TypeRange: hcl.Range{},
	})
}
