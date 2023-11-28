pipeline "nested_pipe_in_child_hcl" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_b" {
        value = step.transform.foo.value
    }
}
