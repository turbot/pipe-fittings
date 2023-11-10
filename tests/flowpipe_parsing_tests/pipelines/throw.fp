pipeline "throw_simple_no_unresolved" {
    step "transform" "one" {
        value = "foo"

        throw {
            if = true
            message = "foo"
        }
    }
}

pipeline "throw_simple_unresolved" {
    step "transform" "one" {
        value = "foo"

        throw {
            if = result.value == "foo"
            message = "foo"
        }
    }
}


pipeline "throw_multiple" {
    step "transform" "one" {
        value = "foo"

        throw {
            if = result.value == "foo"
            message = "foo"
        }

        throw {
            if = true
            message = "bar"
        }

        throw {
            if = result.value == "foo"
            message = "baz"
        }

        throw {
            if = false
            message = "qux"
        }

    }
}

