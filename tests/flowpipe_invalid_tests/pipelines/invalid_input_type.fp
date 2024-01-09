pipeline "pipeline_with_invalid_input_type" {
  step "input" "bad_type" {
    type   = "not_valid"
    prompt = "hello"
  }
}