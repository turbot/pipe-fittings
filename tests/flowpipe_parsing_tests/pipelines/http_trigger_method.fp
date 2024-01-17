pipeline "simple_with_trigger" {
  description = "simple pipeline that will be referred to by a trigger"

  step "transform" "simple_echo" {
    value = "foo bar"
  }
}

trigger "http" "trigger_without_method_block" {
  enabled  = true
  pipeline = pipeline.simple_with_trigger

  args = {
    param_one     = "one"
    param_two_int = 2
  }

  execution_mode = "synchronous"
}
