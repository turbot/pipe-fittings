package modconfig

var AppSpecificNewResourceMapsFunc func(mod *Mod, sourceMaps ...ModResources) ModResources

func NewResourceMaps(mod *Mod, sourceMaps ...ModResources) ModResources {
	if AppSpecificNewResourceMapsFunc == nil {
		panic("AppSpecificNewResourceMapsFunc must be set during app initialization")
	}
	return AppSpecificNewResourceMapsFunc(mod, sourceMaps...)
}
