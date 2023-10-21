test:
	go clean -testcache
	RUN_MODE=TEST_ES go test  $$(go list ./... | grep -v /internal/es/test) -timeout 60s -v

