mod "subtype" {

}

pipeline "subtype" {
    param "conn" {
        type    = string
        default = "example"
        subtype = connection.aws
    }

    param "list_of_conns" {
        type = list(string)
        subtype = list(connection.aws)
    }

    param "conn_generic" {
        type = string
        default = "example"
        subtype = connection
    }

    param "list_of_conns_generic" {
        type = list(string)
        subtype = list(connection)
    }
}
