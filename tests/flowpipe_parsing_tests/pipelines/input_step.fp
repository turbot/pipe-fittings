integration "slack" "integrated_app" {
  token = "abcde"
}

pipeline "pipeline_with_input" {
  step "input" "input" {
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

    prompt = "Choose an option:"

    notify {
      integration = integration.slack.integrated_app
      channel     = param.channel
    }

  }
}
