pipeline "missing_integration_test" {

  step "input" "invalid_notify" {

    prompt = "Are you sure?"

    notify {
      channel = "#general"
    }
  }
}
