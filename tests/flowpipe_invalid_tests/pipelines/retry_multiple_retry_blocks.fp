pipeline "retry_multiple_retry_blocks" {

    step "transform" "one" {
        value = "foo"

        retry {
            retries = 3
        }

        retry {
            retries = 4
        }        
    }
}
