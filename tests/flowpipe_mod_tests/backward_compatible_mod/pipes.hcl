pipeline "parent_pipeline_hcl" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_b" {
        value = step.transform.foo.value
    }
}


pipeline "parent_pipeline_hcl_b" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_b" {
        value = step.transform.foo.value
    }
}
