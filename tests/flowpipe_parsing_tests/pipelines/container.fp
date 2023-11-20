pipeline "pipeline_step_container" {

  description = "Container step test pipeline"

  step "container" "container_test1" {
    image   = "test/image"
    cmd     = ["foo", "bar"]
    timeout = 60
    env = {
      ENV_TEST = "hello world"
    }
  }
}

pipeline "pipeline_step_with_param" {

  description = "Container step test pipeline"

  param "region" {
    description = "The name of the region."
    type        = string
    default     = "ap-south-1"
  }

  param "image" {
    description = "The name of the image."
    type        = string
    default     = "test/image"
  }

  param "cmd" {
    description = "The list of the commands to be run."
    type        = list(string)
    default     = ["foo", "bar"]
  }

  param "timeout" {
    description = "The timeout of the container run."
    type        = number
    default     = 120
  }

  step "container" "container_test1" {
    image   = param.image
    cmd     = param.cmd
    timeout = param.timeout
    env = {
      REGION = param.region
    }
  }
}
