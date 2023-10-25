integration "email" "email_integration" {
  smtp_host = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username = "baz bar foo"
}

pipeline "approval" {
  step "input" "input" {
    notify {
      integration = integration.email.email_integration
      channel = "foo"
    }
  }
}
