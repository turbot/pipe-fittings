pipeline "parent_pipeline_hcl_nested" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_b" {
        value = step.transform.foo.value
    }
}
