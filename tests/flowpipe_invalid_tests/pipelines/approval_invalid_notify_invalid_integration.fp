pipeline "invalid_integration_test" {

  step "input" "invalid_notify" {

    prompt = "Are you sure?"

    notify {
      integration = integration.slack.missing_slack_integration
      channel     = "#general"
    }
  }
}
