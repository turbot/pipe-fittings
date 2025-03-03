package ociinstaller

import (
	"context"
)

type pluginInstallConfig struct {
	skipConfigFile  bool
	getMetadataFunc func(context.Context, string) (*map[string][]string, error)
}

type PluginInstallOption = func(config *pluginInstallConfig)

func WithSkipConfig(skipConfigFile bool) PluginInstallOption {
	return func(o *pluginInstallConfig) {
		o.skipConfigFile = skipConfigFile
	}
}

func WithGetMetadataFunc(getMetadataFunc func(context.Context, string) (*map[string][]string, error)) PluginInstallOption {
	return func(o *pluginInstallConfig) {
		o.getMetadataFunc = getMetadataFunc
	}
}
