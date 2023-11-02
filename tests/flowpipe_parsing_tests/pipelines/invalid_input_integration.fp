pipeline "approval_with_invalid_slack" {
  step "input" "input" {

    notifies = [
      {
        integration = integration.slack.test_app
        to          = "testabc@example.com"
      }
    ]
  }
}
