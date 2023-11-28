mod "mod_child_b" {
  title = "Child Mod b"
}

pipeline "this_pipeline_is_in_the_child" {
    step "transform" "foo" {
        value = "foo"
    }

    output "foo_b" {
        value = step.transform.foo.value
    }
}
