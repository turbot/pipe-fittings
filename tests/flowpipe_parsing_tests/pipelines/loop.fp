pipeline "simple_loop" {
      
    step "echo" "repeat" {
        text  = "iteration"
        numeric = 1

        loop {
            until = result.numeric > 5
            numeric = result.numeric + 1
        }
    }
}

pipeline "simple_http_loop" {
      
  step "http" "list_workspaces" {
    url    = "https://latestpipe.turbot.io/api/v1/org/latesttank/workspace/?limit=3"
    method = "get"

    request_headers = {
      Content-Type  = "application/json"
      Authorization = "Bearer ${param.pipes_token}"
    }

    loop {
      until  = result.response_body.next_token != null
      url = "https://latestpipe.turbot.io/api/v1/org/latesttank/workspace/?limit=3&next_token=${result.response_body.next_token}"
    }
  }
}


pipeline "loop_depeneds_on_another_step" {
      
    step "echo" "base" {
        numeric = 5
    }
    
    step "echo" "repeat" {
        text  = "iteration"
        numeric = 1

        loop {
            until = result.numeric > 5
            numeric = result.numeric + step.echo.base.numeric + 1
        }
    }
}
