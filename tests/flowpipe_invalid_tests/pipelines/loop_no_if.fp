pipeline "simple_loop" {
      
    step "echo" "repeat" {
        text  = "iteration"
        numeric = 1

        loop {
            numeric = result.numeric + 1
        }
    }
}
