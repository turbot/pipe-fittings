mod "custom_type_three" {

}

pipeline "custom_type_two" {
    param "conn" {
        default = connection.aws.example
        type = connection.aws
    }

    param "list_of_conns" {
        
        default = [
            connection.aws.example,
            connection.aws.example_2,
            connection.aws.example_3
        ]
        type = list(connection.aws)
    }

    param "conn_generic" {
        type = connection
    }

    param "list_of_conns_generic" {
        type = list(connection)
    }

    step "transform" "echo" {
        value = param.conn.secret_key
    }
}
