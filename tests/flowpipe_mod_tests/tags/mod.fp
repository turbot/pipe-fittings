mod "tags" {

}

pipeline "with_tags" {
    title = "tags on pipeline"

    tags = {
        "tag1" = "value1"
        "tag2" = "value2"
    }
}