package modconfig

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// BuildModDependencyPath converts a mod dependency name of form github.com/turbot/steampipe-mod-m2
// and a DependencyVersion into a dependency path of form:
// - github.com/turbot/steampipe-mod-m2@v1.0.0
// - github.com/turbot/steampipe-mod-m2#branch
// - github.com/turbot/steampipe-mod-m2:filepath
// This represents the relative path the dependency will be installed at underneath the mods directory
func BuildModDependencyPath(dependencyName string, version *DependencyVersion) string {
	if version == nil {
		// not expected
		return dependencyName
	}

	switch {
	case version.Tag != "":
		return fmt.Sprintf("%s@%s", dependencyName, version.Tag)
	case version.Version != nil:
		return fmt.Sprintf("%s@v%s", dependencyName, version.Version.String())
	case version.Branch != "":
		return fmt.Sprintf("%s#%s", dependencyName, version.Branch)
	case version.FilePath != "":
		// for filepath, we do not use DependencyPath - just return the filepath
		return version.FilePath
	}
	slog.Warn("one of version, branch or file path must be set")
	return dependencyName
}

// BuildModBranchDependencyPath converts a mod dependency name of form github.com/turbot/steampipe-mod-m2
// and a branch into a dependency path of form github.com/turbot/steampipe-mod-m2#branch
func BuildModBranchDependencyPath(dependencyName string, branchName string) string {
	if branchName == "" {
		// not expected
		return dependencyName
	}

	return fmt.Sprintf("%s#%s", dependencyName, branchName)
}

// ParseModDependencyPath converts a mod depdency path of form github.com/turbot/steampipe-mod-m2@v1.0.0
// into the dependency name (github.com/turbot/steampipe-mod-m2) and version
func ParseModDependencyPath(fullName string) (string, *DependencyVersion, error) {
	switch {
	// is this a version constraint
	case strings.Contains(fullName, "@"):
		// split to get the name and version
		parts := strings.Split(fullName, "@")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid mod full name %s", fullName)
			return "", nil, err
		}
		var modVersion *DependencyVersion
		modDependencyName := parts[0]
		versionString := parts[1]
		version, err := semver.NewVersion(versionString)
		if err != nil {
			// if we failed to parse as a semver, treat as a tag
			modVersion = &DependencyVersion{
				Tag: versionString,
			}
		} else {
			// NOTE: we expect the version to be in format 'vx.x.x', i.e. a semver with a preceding v
			if !strings.HasPrefix(versionString, "v") || err != nil {
				return "", nil, fmt.Errorf("mod file %s has invalid version", fullName)
			}
			modVersion = &DependencyVersion{
				Version: version,
			}
		}
		return modDependencyName, modVersion, nil

		// branch constraint
	case strings.Contains(fullName, "#"):
		// split to get the name and branch
		parts := strings.Split(fullName, "#")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid mod full name %s", fullName)
			return "", nil, err
		}
		modDependencyName := parts[0]
		branchName := parts[1]
		modVersion := &DependencyVersion{
			Branch: branchName,
		}
		return modDependencyName, modVersion, nil

	}

	return fullName, nil, nil
}
