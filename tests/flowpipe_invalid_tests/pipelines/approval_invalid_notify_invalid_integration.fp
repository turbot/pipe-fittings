pipeline "invalid_integration_test" {

  step "input" "invalid_notify" {
    type = "button"

    prompt = "Are you sure?"

    notify {
      integration = integration.slack.missing_slack_integration
      channel     = "#general"
    }
  }
}
