mod "local" {

}

integration "slack" "my_slack_app" {
  token = var.slack_token

  # optional - if you want to verify the source
  signing_secret = "Q#$$#@#$$#W"
}

integration "slack" "my_slack_app_two" {
  token = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret = "Q#$$#@#$$#W"
}

integration "email" "email_integration" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
}

// pipeline "approval_with_runtime_param" {

//   param "channel_name" {
//     type = string
//   }

//   step "input" "input" {
//     token = "remove this after integrated"
//     notify {
//       integration = integration.slack.my_slack_app
//       channel = param.channel_name
//     }
//   }
// }

pipeline "approval_with_variables" {

  step "input" "input" {
    type = "button"

    notify {
      integration = integration.slack.my_slack_app
      channel     = var.channel_name
    }
  }
}

variable "channel_name" {
  type    = string
  default = "bar"
}

variable "slack_token" {
  type = string
}
