pipeline "container_step_unimplemented_source" {

  description = "Container step with unimplemented source attribute"

  step "container" "source_test" {
    source = "abc"
    cmd    = []
  }
}
