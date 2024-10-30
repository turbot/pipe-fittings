package modinstaller

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

// updates the 'require' block in 'mod.sp'
func (i *ModInstaller) updateModFile() error {
	contents, err := i.loadModFileBytes()
	if err != nil {
		return err
	}

	oldRequire := i.oldRequire
	newRequire := i.workspaceMod.Require

	// fill these requires in with empty requires
	// so that we don't have to do nil checks everywhere
	// from here on out - if it's empty - it's nil

	if oldRequire == nil {
		// use an empty require as the old requirements
		oldRequire = modconfig.NewRequire()
	}
	if newRequire == nil {
		// use a stub require instance
		newRequire = modconfig.NewRequire()
	}

	changes := EmptyChangeSet()

	if i.shouldDeleteRequireBlock(oldRequire, newRequire) {
		changes = i.buildChangeSetForRequireDelete(oldRequire, newRequire)
	} else if i.shouldCreateRequireBlock(oldRequire, newRequire) {
		changes = i.buildChangeSetForRequireCreate(oldRequire, newRequire)
	} else if !newRequire.Empty() {
		changes = i.calculateChangeSet(oldRequire, newRequire)
	}

	if len(changes) == 0 {
		// nothing to do here
		return nil
	}

	contents.ApplyChanges(changes)
	contents.Apply(hclwrite.Format)
	// strip blank lines
	modData := []byte(helpers.TrimBlankLines(string(contents.Bytes())))

	return os.WriteFile(i.workspaceMod.GetFilePath(), modData, 0644) //nolint:gosec // TODO: check file permission
}

// loads the contents of the mod.sp file and wraps it with a thin wrapper
// to assist in byte sequence manipulation
func (i *ModInstaller) loadModFileBytes() (*ByteSequence, error) {
	modFileBytes, err := os.ReadFile(i.workspaceMod.GetFilePath())
	if err != nil {
		return nil, err
	}
	return NewByteSequence(modFileBytes), nil
}

func (i *ModInstaller) shouldDeleteRequireBlock(oldRequire *modconfig.Require, newRequire *modconfig.Require) bool {
	return newRequire.Empty() && !oldRequire.Empty()
}

func (i *ModInstaller) shouldCreateRequireBlock(oldRequire *modconfig.Require, newRequire *modconfig.Require) bool {
	// NOTE;: only create a new require block if there is NO require block currently
	// if there is an mepty require block we just update it
	// we can detect no require block by examining the TypeRange
	currentModHasRequireBlock := oldRequire.TypeRange.Start.Byte != 0

	return !newRequire.Empty() && !currentModHasRequireBlock
}

func (i *ModInstaller) buildChangeSetForRequireDelete(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	return NewChangeSet(&Change{
		Operation:   Delete,
		OffsetStart: oldRequire.TypeRange.Start.Byte,
		OffsetEnd:   oldRequire.DeclRange.End.Byte,
	})
}

func (i *ModInstaller) buildChangeSetForRequireCreate(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	// if the new require is not empty, but the old one is
	// add a new require block with the new stuff
	// by generating the HCL string that goes in
	f := hclwrite.NewEmptyFile()

	var body *hclwrite.Body
	var insertOffset int

	// we don't have a require block at all
	// let's create one to append to
	body = f.Body().AppendNewBlock("require", nil).Body()
	insertOffset = i.workspaceMod.GetDeclRange().End.Byte - 1

	for _, mvc := range newRequire.Mods {
		newBlock := i.createNewModRequireBlock(mvc)
		body.AppendBlock(newBlock)
	}

	// prefix and suffix with new lines
	// this is so that we can handle empty blocks
	// which do not have newlines
	buffer := bytes.NewBuffer([]byte{'\n'})
	buffer.Write(f.Bytes())
	buffer.WriteByte('\n')

	return NewChangeSet(&Change{
		Operation:   Insert,
		OffsetStart: insertOffset,
		Content:     buffer.Bytes(),
	})
}

func (i *ModInstaller) calculateChangeSet(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	if oldRequire.Empty() && newRequire.Empty() {
		// both are empty
		// nothing to do
		return EmptyChangeSet()
	}
	// calculate the changes
	uninstallChanges := i.calcChangesForUninstall(oldRequire, newRequire)
	installChanges := i.calcChangesForInstall(oldRequire, newRequire)
	updateChanges := i.calcChangesForUpdate(oldRequire, newRequire)

	return MergeChangeSet(
		uninstallChanges,
		installChanges,
		updateChanges,
	)
}

// creates a new "mod" block which can be written as part of the "require" block in mod.sp
func (i *ModInstaller) createNewModRequireBlock(modVersion *modconfig.ModVersionConstraint) *hclwrite.Block {
	modRequireBlock := hclwrite.NewBlock("mod", []string{modVersion.Name})
	if modVersion.BranchName != "" {
		modRequireBlock.Body().SetAttributeValue("branch", cty.StringVal(modVersion.BranchName))
	}
	if modVersion.FilePath != "" {
		modRequireBlock.Body().SetAttributeValue("path", cty.StringVal(modVersion.FilePath))
	}
	if modVersion.Tag != "" {
		modRequireBlock.Body().SetAttributeValue("tag", cty.StringVal(modVersion.Tag))
	} else if modVersion.VersionString != "" {
		modRequireBlock.Body().SetAttributeValue("version", cty.StringVal(modVersion.VersionString))
	}
	return modRequireBlock
}

// calculates changes required in mod.sp to reflect uninstalls
func (i *ModInstaller) calcChangesForUninstall(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	changes := ChangeSet{}
	for _, requiredMod := range oldRequire.Mods {
		// check if this mod is still a dependency
		if modInNew := newRequire.GetModDependency(requiredMod.Name); modInNew == nil {
			changes = append(changes, &Change{
				Operation:   Delete,
				OffsetStart: requiredMod.DefRange.Start.Byte,
				OffsetEnd:   requiredMod.BodyRange.End.Byte,
			})
		}
	}
	return changes
}

// calculates changes required in mod.sp to reflect new installs
func (i *ModInstaller) calcChangesForInstall(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	modsToAdd := []*modconfig.ModVersionConstraint{}
	for _, requiredMod := range newRequire.Mods {
		if modInOld := oldRequire.GetModDependency(requiredMod.Name); modInOld == nil {
			modsToAdd = append(modsToAdd, requiredMod)
		}
	}

	if len(modsToAdd) == 0 {
		// an empty changeset
		return ChangeSet{}
	}

	// create the HCL serialization for the mod blocks which needs to be placed
	// in the require block
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	for _, modToAdd := range modsToAdd {
		rootBody.AppendBlock(i.createNewModRequireBlock(modToAdd))
	}

	return ChangeSet{
		&Change{
			Operation:   Insert,
			OffsetStart: oldRequire.DeclRange.End.Byte - 1,
			Content:     f.Bytes(),
		},
	}
}

// calculates the changes required in mod.sp to reflect updates
func (i *ModInstaller) calcChangesForUpdate(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	changes := ChangeSet{}
	for _, oldRequiredMod := range oldRequire.Mods {
		newRequiredMod := newRequire.GetModDependency(oldRequiredMod.Name)
		if newRequiredMod == nil {
			continue
		}
		// requiredMod.VersionRange contains the locaiton of the current version/tag/branch/path field
		// this field will be replaced with the new value
		var content []byte
		// content will depdend on which property exists in the newRequiredMod
		switch {
		case newRequiredMod.VersionString != "":
			if newRequiredMod.VersionString != oldRequiredMod.VersionString {
				content = []byte(fmt.Sprintf("version = \"%s\"", newRequiredMod.VersionString))
			}
		case newRequiredMod.BranchName != "":
			if newRequiredMod.BranchName != oldRequiredMod.BranchName {
				content = []byte(fmt.Sprintf("branch = \"%s\"", newRequiredMod.BranchName))
			}
		case newRequiredMod.FilePath != "":
			if newRequiredMod.FilePath != oldRequiredMod.FilePath {
				content = []byte(fmt.Sprintf("file_path = \"%s\"", newRequiredMod.FilePath))
			}
		case newRequiredMod.Tag != "":
			if newRequiredMod.Tag != oldRequiredMod.Tag {
				content = []byte(fmt.Sprintf("tag = \"%s\"", newRequiredMod.Tag))
			}
		}
		// has anything changed?
		if len(content) > 0 {
			changes = append(changes, &Change{
				Operation:   Replace,
				OffsetStart: oldRequiredMod.VersionRange.Start.Byte,
				OffsetEnd:   oldRequiredMod.VersionRange.End.Byte,
				Content:     content,
			})
		}
	}

	return changes
}
