package backend

// SetDuckDbExtensionSetupForTest replaces the DuckDB extension setup, which installs extensions from the
// network, and returns a function that puts it back. It lets a test in package backend_test, which can
// import parse, run without the network.
func SetDuckDbExtensionSetupForTest(setup func(exec func(query string) error) error) (restore func()) {
	prev := duckDbExtensionSetup
	duckDbExtensionSetup = setup
	return func() { duckDbExtensionSetup = prev }
}
