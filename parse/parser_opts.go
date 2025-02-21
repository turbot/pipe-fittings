package parse

type ParseHclConfig struct {
	disableTemplateForProperties []string
	escapeBackticks              bool
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

// WithEscapeBackticks is an option to specify whether the grok function should be applied to the hcl file
// this is used to escape grok expressions in the hcl file
func WithEscapeBackticks(applyGrokFunction bool) ParseHclOpt {
	return func(c *ParseHclConfig) {
		c.escapeBackticks = applyGrokFunction
	}
}
