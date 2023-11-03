pipeline "simple_loop" {
      
    step "echo" "repeat" {
        text  = "iteration"
        numeric = 1

        loop {
            until = result.numeric > 5
            baz = result.numeric + 1
        }
    }
}
