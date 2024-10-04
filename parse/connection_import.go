package parse

import (
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"path"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/zclconf/go-cty/cty"
)

func DecodeConnectionImport(configPath string, block *hcl.Block) (*modconfig.ConnectionImport, hcl.Diagnostics) {

	if len(block.Labels) != 1 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid connection_import block - expected 1 label, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	connectionImportName := block.Labels[0]

	connectionImport := modconfig.NewConnectionImport(block)
	if connectionImport == nil {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid connection_import '%s'", connectionImportName),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}
	_, r, diags := block.Body.PartialContent(&hcl.BodySchema{})
	if len(diags) > 0 {
		return nil, diags
	}

	body := r.(*hclsyntax.Body)

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: make(map[string]cty.Value),
	}

	diags = DecodeHclBody(body, evalCtx, nil, connectionImport)
	if len(diags) > 0 {
		return nil, diags
	}

	// moreDiags := connection.Validate()
	// if len(moreDiags) > 0 {
	// 	diags = append(diags, moreDiags...)
	// }

	return connectionImport, diags
}

func ResolveConnectionImportSource(source *string) ([]string, error) {
	var filePaths []string

	// check whether sourcePath is a glob with a root location which exists in the file system
	localSourcePath, glob, err := filehelpers.GlobRoot(*source)
	if err != nil {
		return nil, err
	}
	// if we managed to resolve the sourceDir, treat this as a local path
	if localSourcePath != "" {
		if localSourcePath == "" && glob == "" {
			return filePaths, nil
		}

		// if localSourcePath and glob is same, it indicates that no glob patterns are defined in source
		// determine whether the target is a file or folder
		if localSourcePath == glob {
			// if the path referred a file, then return localSourcePath directly
			if filehelpers.FileExists(localSourcePath) {
				filePaths = append(filePaths, localSourcePath)
				return filePaths, nil
			}
			// must be a folder, append '*' to the glob explicitly, to match all files in that folder.
			glob = path.Join(glob, "*")
		}

		opts := &filehelpers.ListOptions{
			Flags:   filehelpers.AllRecursive,
			Include: []string{glob},
		}

		resolvedFilePaths, err := filehelpers.ListFiles(localSourcePath, opts)
		if err != nil {
			return nil, err
		}
		filePaths = append(filePaths, resolvedFilePaths...)
	}

	return filePaths, nil
}
