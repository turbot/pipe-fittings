pipeline "retry_simple" {

    step "transform" "one" {
        value = "foo"

        retry {
            retries = 3
        }
    }
}

pipeline "retry_with_if" {

    step "transform" "one" {
        value = "foo"

        retry {
            if = result.value == "foo"
            retries = 5
        }
    }
}