package parse

type ParseHclConfig struct {
	disableTemplateForProperties []string
}
type ParseHclOpt func(*ParseHclConfig)

// WithDisableTemplateForProperties is an option to specify the properties for which
// hcl template expression parsing should be disabled
// this is used for preprties which will include a grok path which include the template opening chars '%{'
// if we do not escape the '%{' for these properties, they will fail ot parse
func WithDisableTemplateForProperties(properties []string) ParseHclOpt {
	return func(c *ParseHclConfig) {
		c.disableTemplateForProperties = properties
	}
}
