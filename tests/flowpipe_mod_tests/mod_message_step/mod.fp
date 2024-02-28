mod "mod_message_step" {

}

pipeline "message_step_one" {

    step "message" "hello" {
        notifier = notifier.default
        text = "Hello World"
    }
    
    output "val" {
        value = "Hello World!"
    }
}
