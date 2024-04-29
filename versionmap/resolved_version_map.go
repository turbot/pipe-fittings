package versionmap

// InstalledVersionMap represents a map of ResolvedVersionConstraint, keyed by dependency name
type InstalledVersionMap map[string]*InstalledModVersion

func (m InstalledVersionMap) AddResolvedVersion(constraint *InstalledModVersion) {
	m[constraint.Name] = constraint
}

func (m InstalledVersionMap) Remove(name string) {
	delete(m, name)
}

//// ToVersionListMap converts this map into a ResolvedVersionListMap
//func (m InstalledVersionMap) ToVersionListMap() ResolvedVersionListMap {
//	res := make(ResolvedVersionListMap, len(m))
//	for k, v := range m {
//		res.Add(k, v)
//	}
//	return res
//}
