
integration "slack" "my_slack_app" {
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

pipeline "approval_with_notifies" {

  param "slack_integration" {
    default = true
  }

  param "slack_channel" {
    default = "foo"
  }

  step "input" "input" {
    type = "button"
    option "test" {}

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = "foo"
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

pipeline "approval_with_notifies_depend_another_step" {

  param "slack_integration" {
    default = true
  }

  param "slack_channel" {
    default = "foo"
  }

  step "transform" "echo" {
    value = "some val"
  }
  step "input" "input" {
    type = "button"
    option "test" {}

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = step.transform.echo.value
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
    type = "button"
    option "test" {}

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

pipeline "approval_with_invalid_email_notifies" {
  step "input" "input" {
    type = "button"
    option "test" {}

    notifies = [
      {
        integration = integration.slack.my_slack_app
        channel     = "#general"
      },
      {
        integration = integration.email.email_integration
        channel     = "#general"
      }
    ]
  }
}

pipeline "approval_with_invalid_email" {
  step "input" "input" {
    type = "button"
    option "test" {}

    notifies = [
      {
        integration = integration.email.email_integration
        channel     = "#general"
      }
    ]
  }
}

pipeline "approval_with_invalid_slack" {
  step "input" "input" {
    type = "button"
    option "test" {}

    notifies = [
      {
        integration = integration.slack.my_slack_app
        to          = "testabc@example.com"
      }
    ]
  }
}
