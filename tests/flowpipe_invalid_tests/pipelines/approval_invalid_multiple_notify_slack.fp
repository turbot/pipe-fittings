integration "slack" "slack_dummy_app1" {
  token = "xoxp-lhfahsfhlah035702hfkwsf"
}

integration "slack" "slack_dummy_app2" {
  token = "xoxp-lhfahsfhlah035702hfkwsf"
}

pipeline "approval_invalid_multiple_notify" {

  step "input" "invalid_multiple_notify" {

    prompt = "Are you sure?"

    notify {
      integration = integration.slack.slack_dummy_app1
      channel     = "#general"
    }

    notify {
      integration = integration.slack.slack_dummy_app2
      to          = "test@example.com"
    }
  }
}
