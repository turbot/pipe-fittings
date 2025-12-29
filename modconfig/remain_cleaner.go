package modconfig

// RemainCleaner is implemented by resources that hold HCL Remain fields.
// Calling ClearRemain releases the HCL AST memory after parsing is complete.
type RemainCleaner interface {
	// ClearRemain clears the Remain hcl.Body field to free memory
	ClearRemain()
}

// ClearAllRemain clears Remain fields from all resources in a mod.
// This should be called after parsing is complete to free HCL AST memory.
func ClearAllRemain(mod *Mod) {
	if mod == nil {
		return
	}

	// Clear the mod's own Remain fields
	mod.ClearRemain()

	// Walk all resources and clear their Remain fields
	_ = mod.WalkResources(func(resource HclResource) (bool, error) {
		if cleaner, ok := resource.(RemainCleaner); ok {
			cleaner.ClearRemain()
		}
		return true, nil
	})
}
