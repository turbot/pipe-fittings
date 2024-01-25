

integration "slack" "my_slack_app" {
  token = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret = "Q#$$#@#$$#W"
}

integration "slack" "my_slack_app_two" {
  token = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret = "Q#$$#@#$$#W"
}

integration "email" "email_integration" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
  from            = "test@test.com"
}

integration "email" "email_integration_two" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
  from            = "test@test.com"
}

pipeline "approval_with_notifies_and_multiple_notify" {

  param "slack_integration" {
    default = true
  }

  param "slack_channel" {
    default = "foo"
  }

  step "input" "input" {
    type = "button"

    notify {
      integration = integration.slack.my_slack_app
      channel     = "foo"
    }

    notify {
      integration = integration.email.email_integration_two
      to          = "bob.loblaw@bobloblawlaw.com"
    }

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = "foo"
      },
      {
        integration = integration.email.email_integration
        to          = "bob.loblaw@bobloblawlaw.com"
      }
    ]
  }
}
