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