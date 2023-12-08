mod "invalid_cred_types" {
  title = "invalid_cred_types"
}

pipeline "with_invalid_cred_type_static" {

  param "cred" {
    type    = string
    default = "default"
  }

  step "transform" "test_creds" {
    value = credential.foo[param.cred].token
  }
}
