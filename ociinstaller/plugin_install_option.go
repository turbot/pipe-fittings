package ociinstaller

type pluginInstallConfig struct {
	skipConfigFile  bool
	getMetadataFunc func() (*map[string][]string, error)
}

type PluginInstallOption = func(config *pluginInstallConfig)

func WithSkipConfig(skipConfigFile bool) PluginInstallOption {
	return func(o *pluginInstallConfig) {
		o.skipConfigFile = skipConfigFile
	}
}

func WithGetMetadataFunc(getMetadataFunc func() (*map[string][]string, error)) PluginInstallOption {
	return func(o *pluginInstallConfig) {
		o.getMetadataFunc = getMetadataFunc
	}
}
