integration "slack" "slack_with_token" {
  token = "xoxp-abc123"
}

integration "slack" "slack_with_token_and_secret" {
  token = "xoxp-abc123"
  signing_secret = "W&EYrf78rqwf"
}

integration "slack" "slack_with_webhook" {
  webhook_url = "https://hooks.slack.com/services/T1234567890/B0123456789/XXXXXXXXXX"
}

integration "email" "email_min" {
  smtp_host = "smtp.host.tld"
  from      = "turbie@flowpipe.io"
}

integration "email" "email_with_all" {
  smtp_host     = "123.456.789.000"
  smtp_tls      = "auto"
  smtp_port     = 25
  smtps_port    = 587
  smtp_username = "turbie"
  smtp_password = "some_password_here"

  from              = "turbie@flowpipe.io"
  default_recipient = "user@test.tld"
  default_subject   = "Flowpipe: Action Required"
}