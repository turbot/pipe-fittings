pipeline "http_step" {
  step "http" "send_to_slack" {
    url         = "https://myapi.com/vi/api/do-something"
    method      = "post"
    insecure    = true
    ca_cert_pem = "test"
    timeout     = 2000

    request_body = jsonencode({
      name = "turbie"
      app  = "flowpipe"
    })

    request_headers = {
      Accept     = "application/json"
      User-Agent = "flowpipe" // check - is this the syntax with dash in a key name???
    }
  }
}

pipeline "http_step_timeout_unresolved" {

  param "timeout" {
    type    = number
    default = 2000
  }

  step "http" "send_to_slack" {
    url     = "https://myapi.com/vi/api/do-something"
    method  = "post"
    timeout = param.timeout
  }
}
