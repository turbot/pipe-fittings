package ociinstaller

type PluginImage struct {
	BinaryFile         string
	BinaryDigest       string
	BinaryArchitecture string
	DocsDir            string
	ConfigFileDir      string
	LicenseFile        string
}

func (s *PluginImage) Type() ImageType {
	return ImageTypePlugin
}
