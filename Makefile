test:
	go clean -testcache
	RUN_MODE=TEST_ES go test  $$(go list ./...) -timeout 120s

