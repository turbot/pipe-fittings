dashboard "d1"{
    connection_string="postgresql://kai-fjgw:183f-4c14-b21a@kai-fjgw-wkp1.apse1.db.pipes.turbot-stg.com:9193/ya4lwn"
    table {
        title = "pipes 1"
        sql   = "select * from aws_account"

    }
    table {
        title="pipes 1"
        sql   = "select * from aws_account"
        }

}

dashboard "d2"{
    table {
        title = "pipes"
        sql   = "select * from aws_account"
        connection_string="postgresql://kai-fjgw:183f-4c14-b21a@kai-fjgw-wkp1.apse1.db.pipes.turbot-stg.com:9193/ya4lwn"
    }
    table {
        title="local"
        sql   = "select * from aws_account"
        }
}