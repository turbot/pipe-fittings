integration "slack" "integrated_app" {
  token = "abcde"
}

integration "email" "email_integration" {
  smtp_host     = "smtp.gmail.com"
  from          = "gallant_meitner@example.com"
  smtp_port     = 587
  smtp_username = "awesomebob@blahblah.com"
  smtp_password = "HelloBob@2023"
}

pipeline "pipeline_with_input" {
  step "input" "input" {
    type   = "button"
    prompt = "Choose an option:"

    notify {
      integration = integration.slack.integrated_app
      channel     = "#general"
    }
  }
}

pipeline "pipeline_with_unresolved_notify" {

  param "channel" {
    type    = string
    default = "#general"
  }

  step "input" "input" {
    type = "button"
    prompt = "Choose an option:"

    notify {
      integration = integration.slack.integrated_app
      channel     = param.channel
    }

  }
}

pipeline "pipeline_with_email_notify" {

  param "to" {
    type    = string
    default = "awesomebob@blahblah.com"
  }

  step "input" "input" {
    type = "button"
    # prompt = "Choose an option:"

    notify {
      integration = integration.email.email_integration
      to          = param.to
    }

  }
}
