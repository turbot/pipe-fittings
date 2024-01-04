trigger "query" "query_trigger_interval" {
  schedule = "days"
  pipeline = pipeline.simple_with_trigger

  sql = <<EOQ
        select
            access_key_id,
            user_name,
            create_date,
            ctx ->> 'connection_name' as connection
        from aws_iam_access_key
        where create_date < now() - interval '90 days'
    EOQ

  # Only run the pipeline when keys are newly discovered to have expired
  events      = ["insert"]
  primary_key = "access_key_id"

  args = {
    param_one     = "one"
    param_two_int = 2
  }
}

pipeline "simple_with_trigger" {
    description = "simple pipeline that will be referred to by a trigger"

    step "transform" "simple_echo" {
        value = "foo bar"
    }
}
