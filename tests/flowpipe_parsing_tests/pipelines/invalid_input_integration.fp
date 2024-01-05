pipeline "approval_with_invalid_slack" {
  step "input" "input" {
    type = "button"
    notifies = [
      {
        integration = integration.slack.test_app
        to          = "testabc@example.com"
      }
    ]
  }
}
