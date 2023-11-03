package app_specific

// BuildEnv is a function to construct an application specific env var key
func BuildEnv(suffix string) string {
	return EnvAppPrefix + suffix
}
