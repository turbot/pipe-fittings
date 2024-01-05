integration "slack" "slack_dummy_app" {
  token = "xoxp-fhshf2395723hkhfskh"
}

integration "email" "email_dummy_app" {
  from      = "test@example.com"
  smtp_port = 587
  smtp_host = "smtp.gmail.com"
}

pipeline "approval_multiple_notify_pipeline" {

  step "input" "input_multiple_notify" {
    type   = "button"
    prompt = "Are you sure?"

    notify {
      integration = integration.slack.slack_dummy_app
      channel     = "#general"
    }

    notify {
      integration = integration.email.email_dummy_app
      to          = "test@example.com"
    }
  }
}
