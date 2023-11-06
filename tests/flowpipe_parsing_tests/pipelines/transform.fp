pipeline "pipeline_with_transform_step" {

  description = "Pipeline with a valid transform step"

  step "transform" "transform_test" {
    value = "hello world"
  }
}

pipeline "pipeline_with_transform_step_unresolved" {

  description = "Pipeline with a valid transform step (unresolved)"

  param "random_text" {
    type    = string
    default = "hello world"
  }

  step "transform" "transform_test" {
    value = param.random_text
  }
}

pipeline "pipeline_with_transform_step_number_test" {

  description = "Pipeline with a valid transform step with number value"

  step "transform" "transform_test" {
    value = 100
  }
}

pipeline "pipeline_with_transform_step_number_test_unresolved" {

  description = "Pipeline with a valid transform step with number value (unresolved)"

  param "random" {
    type    = number
    default = 1000
  }

  step "transform" "transform_test" {
    value = param.random
  }
}

pipeline "pipeline_with_transform_step_string_list" {

  description = "Pipeline with a valid transform step contains list of strings"

  param "users" {
    type    = list(string)
    default = ["brian", "freddie", "john", "roger"]
  }

  step "transform" "transform_test" {
    for_each = param.users
    value    = "user if ${each.value}"
  }
}

pipeline "pipeline_with_transform_step_number_list" {

  description = "Pipeline with a valid transform step contains list of numbers"

  param "counts" {
    type    = list(string)
    default = [1, 2, 3]
  }

  step "transform" "transform_test" {
    for_each = param.counts
    value    = "counter set to ${each.value}"
  }
}
