package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/versionhelpers"
	"github.com/zclconf/go-cty/cty"
)

const FilePrefix = "file:"

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
	// TODO KAI maybe only export version and populate filepath/branch name after loading
	// the local file location to use
	FilePath string `cty:"file_path" hcl:"file_path,optional"`
	// the branch name to use
	BranchName string `cty:"branch" hcl:"branch,optional"`

	// only one of VersionConstraint, Branch and FilePath will be set
	versionConstraint *versionhelpers.Constraints

	// contains the range of the definition of the mod block
	DefRange hcl.Range
	// contains the range of the body of the mod block
	BodyRange hcl.Range
	// contains the range of the total version field
	VersionRange hcl.Range
}

func NewFilepathModVersionConstraint(mod *Mod) *ModVersionConstraint {
	return &ModVersionConstraint{
		Args:     make(map[string]cty.Value),
		Name:     mod.ShortName,
		FilePath: mod.ModPath,
	}
}

// NewModVersionConstraint creates a new ModVersionConstraint - this is called when installing a mod
func NewModVersionConstraint(modFullName string) (*ModVersionConstraint, error) {
	m := &ModVersionConstraint{
		Args: make(map[string]cty.Value),
	}

	switch {
	case strings.HasPrefix(modFullName, FilePrefix):
		// filepath constraints should be handled separately
		return nil, fmt.Errorf("NewModVersionConstraint does not support file paths - use NewFilepathModVersionConstraint")
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
		// record all the ranges in the source file
		m.DefRange = block.DefRange
		m.BodyRange = block.Body.(*hclsyntax.Body).SrcRange
		// record the range of the version attribute in this structure
		if versionAttribute, ok := block.Body.(*hclsyntax.Body).Attributes["version"]; ok {
			m.VersionRange = versionAttribute.SrcRange
		}
	}

	// only 1 of version, branch or file path should be set
	count := 0
	if m.VersionString != "" {
		count++
	}
	if m.BranchName != "" {
		count++
	}
	if m.FilePath != "" {
		count++
	}
	if count > 1 {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "only one of 'version', 'branch' or 'file_path' should be set",
			Subject:  &m.DefRange,
		}}
	}

	// if a branch name or file path is set, nothing more to do
	if m.BranchName != "" || m.FilePath != "" {
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

	// so there was an error
	return hcl.Diagnostics{&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("invalid mod version %s", m.VersionString),
		Subject:  &m.DefRange,
	}}

}

func (m *ModVersionConstraint) DependencyPath() string {
	if m.HasVersion() {
		return fmt.Sprintf("%s@%s", m.Name, m.VersionString)
	}
	if m.BranchName != "" {
		return fmt.Sprintf("%s#%s", m.Name, m.BranchName)
	}
	if m.FilePath != "" {
		return fmt.Sprintf("%s%s", FilePrefix, m.FilePath)
	}
	return m.Name
}

// HasVersion returns whether the mod has a version specified, or is the latest
// if no version is specified, or the version is "latest", this is the latest version
func (m *ModVersionConstraint) HasVersion() bool {
	return !helpers.StringSliceContains([]string{"", "latest", "*"}, m.VersionString)
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
		return fmt.Sprintf("file:m.FilePath")
	case m.BranchName != "":
		return fmt.Sprintf("branch:%s", m.BranchName)
	}
	panic("one of version, branch or file path must be set")
}
