test:
	go clean -testcache
	RUN_MODE=TEST_ES go test ./... -timeout 120s | grep -v "no test files"