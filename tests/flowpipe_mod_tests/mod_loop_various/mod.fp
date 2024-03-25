mod "test" {

}

pipeline "sleep" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 2
        }
    }
}

pipeline "sleep_2" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 2
            duration = "10s"
        }
    }
}

pipeline "sleep_3" {

    step "sleep" "one" {
        duration = "5s"

        loop {
            until = loop.index > 2
            duration = "${loop.index}s"
        }
    }
}

pipeline "sleep_4" {

    step "sleep" "one" {
        duration = "5"

        loop {
            until = loop.index > 2

            # reference to result used to cause failure in this block, do not remove this test
            duration = "${loop.index}${result.duration}"
        }
    }
}

pipeline "http" {

    step "http" "http" {
        url = "https://foo"

        loop {
            until = loop.index > 2
            url = "https://bar"
        }
    }
}

pipeline "http_2" {

    step "http" "http" {
        url = "https://foo"

        loop {
            until = loop.index > 2
            url = "https://bar/${loop.index}"
        }
    }
}

pipeline "http_3" {

  step "http" "http" {
    url = "http://localhost:7104/special-case"
    method = "post"
    request_body = jsonencode({
      query = "bar"
    })

    loop {
      until = loop.index >= 2
      request_body = replace(result.request_body, "bar", "baz")
    }
  }
}

pipeline "container" {

    step "container" "container" {
        image = "alpine:3.7"

        cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]

        env = {
        FOO = "bar"
        }

        timeout            = 60000 // in ms
        memory             = 128
        memory_reservation = 64
        memory_swap        = 256
        memory_swappiness  = 10
        read_only          = false
        user               = "root"

        loop {
            until = loop.index >= 2
            memory = 150 + loop.index            
        }        
    }
}

pipeline "container_2" {

    step "container" "container" {
        image = "alpine:3.7"

        cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]

        env = {
        FOO = "bar"
        }

        timeout            = 60000 // in ms
        memory             = 128
        memory_reservation = 64
        memory_swap        = 256
        memory_swappiness  = 10
        read_only          = false
        user               = "root"

        loop {
            until = loop.index >= 2
            memory = 150 + loop.index
            cmd = ["a", "b", "c"]
        }        
    }
}

pipeline "container_3" {

    step "container" "container" {
        image = "alpine:3.7"

        cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]
        entrypoint = ["1", "2"]

        env = {
        FOO = "bar"
        }

        timeout            = 60000 // in ms
        memory             = 128
        memory_reservation = 64
        memory_swap        = 256
        memory_swappiness  = 10
        read_only          = false
        user               = "root"

        loop {
            until = loop.index >= 2
            memory = 150 + loop.index
            cmd = ["a", "b", "c"]
            entrypoint = ["1", "2"]
            cpu_shares = 4
        }        
    }
}

pipeline "container_4" {

    step "container" "container" {
        image = "alpine:3.7"

        cmd = [ "sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'" ]
        entrypoint = ["1", "2"]

        env = {
            FOO = "bar"
        }

        timeout            = 60000 // in ms
        memory             = 128
        memory_reservation = 64
        memory_swap        = 256
        memory_swappiness  = 10
        read_only          = false
        user               = "root"

        loop {
            until = loop.index >= 2
            memory = 150 + loop.index
            cmd = ["a", "b", "c"]
            entrypoint = ["1", "2"]
            cpu_shares = 4
            env = {
                bar = "baz"
            }
        }        
    }
}

pipeline "pipeline" {

    step "pipeline" "pipeline" {
        pipeline = pipeline.nested

        args = {
            a = "foo_2"
            c = 44
        }

        loop {
            until = loop.index >= 2
            args = {
                a = "foo_10"
                c = 44 + loop.index
            }
        }
    }
}

pipeline "pipeline_2" {

    step "pipeline" "pipeline" {
        pipeline = pipeline.nested

        args = {
            a = "foo_2"
            c = 44
        }

        loop {
            until = loop.index >= 2
            args = {
                a = "foo_10"
                c = 44
            }
        }
    }
}

pipeline "pipeline_3" {

    step "pipeline" "pipeline" {
        pipeline = pipeline.nested

        args = {
            a = "foo_2"
            c = 44
        }

        loop {
            until = try(result.output.test, null) == null
            args = {
                a = "foo_10"
                c = 44
            }
        }
    }
}

pipeline "nested" {
    param "a" {
        default = "foo"
    }

    param "b" {
        default = "bar"
    }

    param "c" {
        default = 42
    }

    step "transform" "echo" {
        value = "bar"
    }
}

pipeline "query" {

    step "query" "query" {
        sql      = "select * from aws_account"
        database = "postgres://steampipe:@host.docker.internal:9193/steampipe"
        args = [
            "foo"
        ]

        loop {
            until = loop.index >= 2
            args = [
                "bar"
            ]
        }
    }
}

pipeline "query_2" {

    step "query" "query" {
        sql      = "select * from aws_account"
        database = "postgres://steampipe:@host.docker.internal:9193/steampipe"
        args = [
            "foo"
        ]

        loop {
            until = loop.index >= 2
            args = [
                "bar",
                loop.index
            ]
        }
    }
}

