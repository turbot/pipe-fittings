package ociinstaller

const DefaultConfigSchema string = "2020-11-18"

type OciImageConfig interface {
	GetSchemaVersion() string
	SetSchemaVersion(string)
}

type OciConfigBase struct {
	SchemaVersion string `json:"schemaVersion"`
}

func (c *OciConfigBase) GetSchemaVersion() string {
	return c.SchemaVersion
}
func (c *OciConfigBase) SetSchemaVersion(version string) {
	c.SchemaVersion = version
}

type PluginImageConfig struct {
	OciConfigBase
	Plugin struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
	}
}
