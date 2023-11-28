mod "test_mod" {
  title = "my_mod"
}

variable "schedule_default" {
  type        = string
  description = "schedule with default value"
  default     = "5 * * * *"
}

variable "interval_default" {
  type        = string
  description = "interval with default value"
  default     = "weekly"
}

variable "var_one" {
  type        = string
  description = "test variable"
  default     = "this is the value of var_one"
}

# var_two will be overriden in the test
variable "var_two" {
  type        = string
  description = "test variable"
  default     = "default of var_two"
}


# var_three has no default
variable "var_three" {
  type        = string
  description = "test variable"
}

# var_four has no default
variable "var_four" {
  type        = string
  description = "test variable"
}

# var_five has no default
variable "var_five" {
  type        = string
  description = "test variable"
}

# var_six has no default
variable "var_six" {
  type        = string
  description = "test variable"
}

pipeline "one" {
  step "transform" "one" {
    value = "prefix text here and ${var.var_one} and suffix"
  }

  step "transform" "two" {
    value = "prefix text here and ${var.var_two} and suffix"
  }

  step "transform" "three" {
    value = "prefix text here and ${var.var_three} and suffix"
  }

  step "transform" "one_echo" {
    value = "got prefix? ${step.transform.one.value} and again ${step.transform.one.value} and var ${var.var_one}"
  }


  step "transform" "four" {
    value = "using value from locals: ${local.locals_one}"
  }

  step "transform" "five" {
    value = "using value from locals: ${local.locals_two}"
  }

  step "transform" "six" {
    value = "using value from locals: ${local.locals_three.key_two}"
  }

  step "transform" "seven" {
    value = "using value from locals: ${local.locals_three_merge.key_two}"
  }

  step "transform" "eight" {
    value = "using value from locals: ${local.locals_three_merge.key_three}"
  }

  step "transform" "eight" {
    value = "var_four value is: ${var.var_four}"
  }

  step "transform" "nine" {
    value = "var_five value is: ${var.var_five}"
  }

  step "transform" "ten" {
    value = "var_six value is: ${var.var_six}"
  }
}


variable "default_gh_repo" {
  type    = string
  default = "hello-world"
}

pipeline "github_issue" {
  param "gh_repo" {
    type    = string
    default = var.default_gh_repo
  }

  step "http" "get_issue" {
    url = "https://api.github.com/repos/octocat/${param.gh_repo}/issues/2743"
  }
}


pipeline "github_get_issue_with_number" {

  param "github_token" {
    type = string
  }

  param "github_issue_number" {
    type = number
  }

  step "http" "get_issue" {
    title  = "Get details about an issue"
    method = "post"
    url    = "https://api.github.com/graphql"
    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.github_token}"
    }

    request_body = jsonencode({
      query = <<EOM
              query {
                repository(owner: "octocat", name: "hello-world") {
                  issue(number: ${param.github_issue_number}) {
                    id
                    number
                    url
                    title
                    body
                  }
                }
              }
            EOM
    })
  }
}

locals {
  locals_three_merge = merge(local.locals_three, {
    key_three = 33
  })
}

locals {
  locals_one = "value of locals_one"

  locals_two = 10

  locals_three = {
    key_one = "value of key_one"
    key_two = "value of key_two"
  }

  locals_four = ["foo", "bar", "baz"]
}
