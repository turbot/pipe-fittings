mod "test" {

}

pipeline "input_with_loop" {

    step "input" "my_input" {
        prompt   = "Shall we play a game?"
        type     = "select"
        notifier = notifier.default

        option "Tic Tac Toe" {}
        option "Checkers" {}
        option "Global Thermonuclear War" {}

        loop {
            until = loop.index > 2
            notifier = notifier["notifier_{$loop.index}"]
        }
    }
}