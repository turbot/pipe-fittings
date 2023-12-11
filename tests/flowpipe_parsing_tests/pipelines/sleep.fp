pipeline "pipeline_with_sleep" {

  step "sleep" "sleep_duration_string_input" {
    duration = "5s"
  }

  step "sleep" "sleep_duration_integer_input" {
    duration = 2000
  }

}