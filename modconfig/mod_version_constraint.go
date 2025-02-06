package modconfig

import (
	"fmt"
	"slices"
	"strings"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/versionhelpers"
	"github.com/zclconf/go-cty/cty"
)

type VersionConstrainCollection []*ModVersionConstraint

// ModVersionConstraint is a struct to represent a version as specified in a mod require block
type ModVersionConstraint struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name string `cty:"name" hcl:"name,label"`
	// variable values to be set on the dependency mod
	Args map[string]cty.Value `cty:"args"  hcl:"args,optional"`

	// optionally override the database and search path
	Database         *string  `cty:"database" hcl:"database"`
	SearchPath       []string `cty:"search_path" hcl:"search_path,optional"`
	SearchPathPrefix []string `cty:"search_path_prefix" hcl:"search_path_prefix,optional"`
	// the version constraint string
	VersionString string `cty:"version" hcl:"version,optional"`
	// the local file location to use
	FilePath string `cty:"path" hcl:"path,optional"`
	// the branch name to use
	BranchName string `cty:"branch" hcl:"branch,optional"`
	// the (non-version) tag to use
	// populated only if a tag which is not a semver is used
	Tag string `cty:"tag" hcl:"tag,optional"`

	// only one of VersionConstraint, Branch and FilePath will be set
	versionConstraint *versionhelpers.Constraints

	// contains the range of the definition of the mod block
	DefRange hcl.Range
	// contains the range of the body of the mod block
	BodyRange hcl.Range
	// contains the range of the version/branch/tag/path field
	VersionRange hcl.Range
}

func NewFilepathModVersionConstraint(mod *Mod) *ModVersionConstraint {
	return &ModVersionConstraint{
		Args: make(map[string]cty.Value),
		// set name and filepath to the same value
		Name:     mod.ModPath,
		FilePath: mod.ModPath,
	}
}

// NewModVersionConstraint creates a new ModVersionConstraint - this is called when installing a mod
func NewModVersionConstraint(modFullName string) (*ModVersionConstraint, error) {
	m := &ModVersionConstraint{
		Args: make(map[string]cty.Value),
	}
	switch {
	case strings.Contains(modFullName, "@"):
		// try to extract version from name
		segments := strings.Split(modFullName, "@")
		if len(segments) > 2 {
			return nil, fmt.Errorf("invalid mod name %s", modFullName)
		}
		m.Name = segments[0]
		m.VersionString = segments[1]
	case strings.Contains(modFullName, "#"):
		// try to extract branch from name
		segments := strings.Split(modFullName, "#")
		if len(segments) > 2 {
			return nil, fmt.Errorf("invalid mod name %s", modFullName)
		}
		m.Name = segments[0]
		m.BranchName = segments[1]
	case filehelpers.DirectoryExists(modFullName):
		// filepath constraints should be handled separately
		return nil, fmt.Errorf("NewModVersionConstraint does not support file paths - use NewFilepathModVersionConstraint")

	default:
		m.Name = modFullName
	}

	// try to convert version into a semver constraint
	if diags := m.Initialise(nil); diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("failed to initialise version constraint", diags)
	}
	return m, nil
}

// Initialise parses the version and name properties
func (m *ModVersionConstraint) Initialise(block *hcl.Block) hcl.Diagnostics {
	if block != nil {
		diags := m.setRanges(block)
		if diags.HasErrors() {
			return diags
		}
	}

	// only 1 of version, branch, file path or tag should be set
	// Ensure that only one of version, branch, file path, or tag is set.
	fields := []string{m.VersionString, m.BranchName, m.FilePath, m.Tag}
	var activeField string
	for _, field := range fields {
		if field != "" {
			if activeField != "" {
				return hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "only one of 'version', 'branch', 'path', or 'tag' should be set",
					Subject:  &m.DefRange,
				}}
			}
			activeField = field
		}
	}

	// if a branch name or file path is set, nothing more to do
	if m.BranchName != "" || m.FilePath != "" || m.Tag != "" {
		return nil
	}

	// otherwise, if create a version constraint from the version
	// now default the version string to latest
	if m.VersionString == "" || m.VersionString == "latest" {
		m.VersionString = "*"
	}

	// does the version parse as a semver version
	if c, err := versionhelpers.NewConstraint(m.VersionString); err == nil {
		// no error
		m.versionConstraint = c
		return nil
	}

	// if we get here we failed to parse the version string as a semver - treat it as a git tag instead - we will verify it later
	m.Tag = m.VersionString
	// NOTE: clear the version string
	m.VersionString = ""
	return nil
}

func (m *ModVersionConstraint) setRanges(block *hcl.Block) hcl.Diagnostics {
	// record all the ranges in the source file
	m.DefRange = block.DefRange
	m.BodyRange = block.Body.(*hclsyntax.Body).SrcRange
	// record the range of the version/branch/path/tag attribute in this structure
	if versionAttribute, ok := block.Body.(*hclsyntax.Body).Attributes["version"]; ok {
		m.VersionRange = versionAttribute.SrcRange
	} else if branchAttribute, ok := block.Body.(*hclsyntax.Body).Attributes["branch"]; ok {
		m.VersionRange = branchAttribute.SrcRange
	} else if pathAttribute, ok := block.Body.(*hclsyntax.Body).Attributes["path"]; ok {
		m.VersionRange = pathAttribute.SrcRange
	} else if tagAttribute, ok := block.Body.(*hclsyntax.Body).Attributes["tag"]; ok {
		m.VersionRange = tagAttribute.SrcRange
	} else {
		// one of these must be present
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to load mod require block: one of 'version', 'branch', 'path', or 'tag' must be set",
			Subject:  &m.DefRange,
		}}
	}
	return nil
}

func (m *ModVersionConstraint) DependencyPath() string {
	switch {
	case m.HasVersion():
		return fmt.Sprintf("%s@%s", m.Name, m.VersionString)
	case m.Tag != "":
		return fmt.Sprintf("%s@%s", m.Name, m.Tag)
	case m.BranchName != "":
		return fmt.Sprintf("%s#%s", m.Name, m.BranchName)
	case m.FilePath != "":
		return m.FilePath
	default:
		return m.Name
	}
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersionConstraint) HasVersion() bool {
	return !slices.Contains([]string{"", "latest", "*"}, m.VersionString)
}

func (m *ModVersionConstraint) String() string {
	return m.DependencyPath()
}

func (m *ModVersionConstraint) Equals(other *ModVersionConstraint) bool {
	// just check the hcl properties
	return m.Name == other.Name && m.VersionString == other.VersionString
}

func (m *ModVersionConstraint) IsPrerelease() bool {
	return m.versionConstraint != nil && m.versionConstraint.IsPrerelease()
}

func (m *ModVersionConstraint) VersionConstraint() *versionhelpers.Constraints {
	return m.versionConstraint
}

func (m *ModVersionConstraint) OriginalConstraint() any {
	switch {
	case m.VersionString != "":
		return m.VersionString
	case m.FilePath != "":
		return m.FilePath
	case m.BranchName != "":
		return fmt.Sprintf("branch:%s", m.BranchName)
	case m.Tag != "":
		return fmt.Sprintf("branch:%s", m.Tag)
	}
	panic("one of version, branch, file path or tag must be set")
}
