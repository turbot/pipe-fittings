pipeline "invalid_http_request_timeout" {
  step "http" "invalid_request_timeout" {
    url                = "https://myapi.com/vi/api/do-something"
    method             = "post"
    request_timeout_ms = ["1s"]
  }
}
