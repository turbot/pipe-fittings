integration "email" "email_dummy_app1" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
}

integration "email" "email_dummy_app2" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
}

pipeline "approval_invalid_multiple_notify" {

  step "input" "invalid_multiple_notify" {

    prompt = "Are you sure?"

    notify {
      integration = integration.email.email_dummy_app1
      channel     = "#general"
    }

    notify {
      integration = integration.email.email_dummy_app2
      to          = "test@example.com"
    }
  }
}
