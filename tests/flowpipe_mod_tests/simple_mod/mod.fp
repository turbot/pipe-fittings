mod "test_mod" {
  title = "my_mod"

  require {
    flowpipe {
      min_version = "0.1.0"
    }
  }

  tags = {
    foo = "bar"
    green = "day"
  }
}

pipeline "simple" {

  param "foo" {
    type = string
  }

  param "name" {
    type = string
  }

  param "name" {
    type = string
  }

  step "transform" "echo" {
    value = "literal foo"
  }

  step "transform" "eval" {
    value = param.foo
  }

  step "transform" "eval_with_literal" {
    value = "literal ${param.foo} end literal"
  }
 
  step "transform" "name" {
      if = param.name == "foo"
      value = "echo ${param.name}"
  }

  step "http" "list_issues" {
    method = "get"
    url    = "${param.api_base_url}/rest/api/2/search?jql=project=${param.project_key}"
    request_headers = {
      Content-Type = "application/json"
      Foo = "bar"
      Baz = "qux"
    }
  }
}


pipeline "jsonplaceholder_expr" {
    description = "Simple pipeline to demonstrate HTTP post operation."

    step "transform" "method" {
        value = "post"
    }

    param "timeout" {
        type = number
        default = 1000
    }

    param "user_agent" {
        type = string
        default = "flowpipe"
    }

    param "insecure" {
        type = bool
        default = true
    }

    step "http" "http_1" {
        url = "https://jsonplaceholder.typicode.com/posts"

        method = step.transform.method.value

        request_body = jsonencode({
            userId = 12345
            title = ["brian may", "freddie mercury", "roger taylor", "john deacon"]
            nested = {
                brian = {
                    may = "guitar"
                }
                freddie = {
                    mercury = "vocals"
                }
                roger = {
                    taylor = "drums"
                }
                john = {
                    deacon = "bass"
                }
            }
        })

        request_headers = {
            Accept = "*/*"
            Content-Type = "application/json"
            User-Agent = param.user_agent
        }

        insecure = param.insecure
    }

    step "transform" "output" {
        value = step.http.http_1.status_code
    }

    step "transform" "body_json" {
        value = step.http.http_1.response_body
    }

    step "transform" "body_json_nested" {
        value = step.http.http_1.response_body["nested"]["brian"]["may"]
    }

    step "transform" "body_json_loop" {
        for_each = step.http.http_1.response_body["title"]
        value = each.value
    }

    output "foo" {
        value = step.http.http_1.response_body
    }

    output "nested" {
        value = step.transform.body_json_nested.value
    }
}