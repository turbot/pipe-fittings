pipeline "pipeline_with_transform_step" {

  description = "Pipeline with a valid transform step"

  param "random_text" {
    type    = string
    default = "hello world"
  }

  step "transform" "transform_test" {
    value = "hello world"
  }
}
