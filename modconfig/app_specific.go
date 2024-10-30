package modconfig

var AppSpecificNewModResourcesFunc func(mod *Mod, sourceMaps ...ModResources) ModResources

func NewModResources(mod *Mod, sourceMaps ...ModResources) ModResources {
	if AppSpecificNewModResourcesFunc == nil {
		panic("AppSpecificNewModResourcesFunc must be set during app initialization")
	}
	return AppSpecificNewModResourcesFunc(mod, sourceMaps...)
}
