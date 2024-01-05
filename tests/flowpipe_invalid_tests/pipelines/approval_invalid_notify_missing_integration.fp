pipeline "missing_integration_test" {

  step "input" "invalid_notify" {
    type = "button"

    prompt = "Are you sure?"

    notify {
      channel = "#general"
    }
  }
}
