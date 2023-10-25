integration "slack" "my_slack_app" {
  token           = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

pipeline "approval" {
  step "input" "input" {
    notify {
      integration = integration.slack.my_slack_app
      bad_attribute = "foo"
    }
  }
}
