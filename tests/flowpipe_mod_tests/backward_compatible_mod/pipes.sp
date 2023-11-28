pipeline "parent_pipeline_sp" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_b" {
        value = step.transform.foo.value
    }
}

