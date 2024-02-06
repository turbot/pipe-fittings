
mod "mod_with_integration" {
  title = "mod_with_integration"
}

pipeline "approval_with_notifies" {

  param "slack_integration" {
    default = true
  }

  param "slack_channel" {
    default = "foo"
  }

  // step "input" "input" {
  //   type = "button"
  //   option "test" {}

  //   notifies = [
  //     {
  //       integration = integration.slack.my_slack_app
  //       channel     = "foo"
  //       # channel = param.slack_channel
  //       # if      = param.slack_integration == null ? false : true
  //     }
  //   ]
  // }
}