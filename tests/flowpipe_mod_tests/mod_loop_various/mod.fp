mod "test" {

}

pipeline "sleep" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 2
        }
    }
}

pipeline "sleep_2" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 2
            duration = "10s"
        }
    }
}

pipeline "sleep_3" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 2
            duration = "${loop.index}s"
        }
    }
}

pipeline "sleep_4" {

    step "sleep" "one" {
        duration = "5"

        loop {
            until = loop.index > 2

            # reference to result used to cause failure in this block, do not remove this test
            duration = "${loop.index}${result.duration}"
        }
    }
}

pipeline "http" {

    step "http" "http" {
        url = "https://foo"

        loop {
            until = loop.index > 2
            url = "https://bar"
        }
    }
}

pipeline "http_2" {

    step "http" "http" {
        url = "https://foo"

        loop {
            until = loop.index > 2
            url = "https://bar/${loop.index}"
        }
    }
}

pipeline "http_3" {

  step "http" "http" {
    url = "http://localhost:7104/special-case"
    method = "post"
    request_body = jsonencode({
      query = "bar"
    })

    loop {
      until = loop.index >= 2
      request_body = replace(result.request_body, "bar", "baz")
    }
  }
}