pipeline "pipeline_with_option_and_options" {
  step "input" "multiple_option_and_options" {
    type   = "button"
    prompt = "choose one:"

    option "yes" {}
    option "maybe" {}

    options = [
      {
        "value": "no"
      }
    ]
  }
}