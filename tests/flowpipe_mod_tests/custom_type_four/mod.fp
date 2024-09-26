mod "custom_type_four" {
}


variable "conn" {
    type = connection.aws
    default = connection.aws.example
}

variable "list_of_conns" {
    type = list(connection.aws)
    default = [
        connection.aws.example,
        connection.aws.example_2,
        connection.aws.example_3
    ]
}

variable "conn_generic" {
    type = connection
      default = connection.aws.example

}

variable "list_of_conns_generic" {
    type = list(connection)
     default = [
            connection.aws.example,
            connection.aws.example_2,
            connection.aws.example_3
        ]
}


pipeline "a" {

     step "query" "select" {
        database = connection.aws.example
        sql = "SELECT 1"
    }
}

pipeline "b" {

     step "query" "select" {
        database = var.conn
        sql = "SELECT 1"
    }
}
