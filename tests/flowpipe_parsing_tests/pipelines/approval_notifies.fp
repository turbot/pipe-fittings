

integration "slack" "my_slack_app" {
  token = "xoxp-111111"
  # optional - if you want to verify the source
  signing_secret = "Q#$$#@#$$#W"
}

integration "email" "email_integration" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
}

pipeline "approval_with_notifies" {

  param "slack_integration" {
    default = true
  }

  param "slack_channel" {
    default = "foo"
  }

  step "input" "input" {

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = "foo"
        to          = "bob.loblaw@bobloblawlaw.com"
        # channel = param.slack_channel
        # if      = param.slack_integration == null ? false : true
      },
      {
        integration = integration.email.email_integration
        to          = "bob.loblaw@bobloblawlaw.com"
        channel     = "bar"
      }
    ]
  }
}

pipeline "approval_with_notifies_depend_another_step" {

  param "slack_integration" {
    default = true
  }

  param "slack_channel" {
    default = "foo"
  }

  step "echo" "echo" {
    text = "some val"
  }
  step "input" "input" {

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = step.echo.echo.text
        # channel = param.slack_channel
        # if      = param.slack_integration == null ? false : true
      },
      {
        integration = integration.email.email_integration
        to          = "bob.loblaw@bobloblawlaw.com"
      }
    ]
  }
}

pipeline "approval_with_invalid_notifies" {
  step "input" "input" {

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = "#general"
        to          = "bob.loblaw@bobloblawlaw.com"
      },
      {
        integration = integration.email.email_integration
        to          = "bob.loblaw@bobloblawlaw.com"
      }
    ]
  }
}
