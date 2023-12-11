mod "test_mod" {
  title = "my_mod"

  require {
    flowpipe {
      min_version = "0.1.0"
    }
  }

  tags = {
    foo = "bar"
    green = "day"
  }
}

pipeline "simple" {

  param "foo" {
    type = string
  }

  step "transform" "echo" {
    value = "literal foo"
  }

  step "transform" "eval" {
    value = param.foo
  }

  step "transform" "eval_with_literal" {
    value = "literal ${param.foo} end literal"
  }
}
