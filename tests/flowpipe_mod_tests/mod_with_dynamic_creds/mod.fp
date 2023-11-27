mod "mod_with_creds" {
  title = "mod_with_creds"
}


pipeline "cred_aws" {
    param "cred" {
        type    = string
        default = "aws_static"
    }

    step "transform" "aws" {
        value   = credential.aws[param.cred].env
    }

    step "transform" "aws_access_key" {
        value = credential.aws[param.cred].access_key
    }

    output "val" {
        value = step.transform.aws.value
    }

    output "val_access_key" {
        value = step.transform.aws_access_key.value
    }
}
