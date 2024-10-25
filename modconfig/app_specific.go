package modconfig

var AppSpecificNewResourceMapsFunc func(mod ModI, sourceMaps ...ResourceMapsI) ResourceMapsI

func NewResourceMaps(mod ModI, sourceMaps ...ResourceMapsI) ResourceMapsI {
	if AppSpecificNewResourceMapsFunc == nil {
		panic("AppSpecificNewResourceMapsFunc must be set during app initialization")
	}
	return AppSpecificNewResourceMapsFunc(mod, sourceMaps...)
}
