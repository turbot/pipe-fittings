mod "mod_message_step" {

}

pipeline "echo" {

    step "message" "hello" {
        body = "Hello World"
    }
    
    output "val" {
        value = "Hello World!"
    }
}
