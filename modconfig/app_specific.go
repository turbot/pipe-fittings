package modconfig

var AppSpecificNewResourceMapsFunc func(mod *Mod, sourceMaps ...ResourceMapsI) ResourceMapsI

func NewResourceMaps(mod *Mod, sourceMaps ...ResourceMapsI) ResourceMapsI {
	if AppSpecificNewResourceMapsFunc == nil {
		panic("AppSpecificNewResourceMapsFunc must be set during app initialization")
	}
	return AppSpecificNewResourceMapsFunc(mod, sourceMaps...)
}
