http_server server0 {
  address = "127.0.0.1:2345"
  timeout = "10s"

  endpoint {
    path = "/reset_dms/${env.DMS_SECRET}"
    send_to = dms0
  }

  endpoint {
    methods = ["GET"]
    path = "/forward"
    send_to = dms0
  }

  monitoring_endpoint {
    path = "/metrics"
  }
}

dead_mans_switch dms0 {
  timeout = "10s"
  send_to = [
    dms0_request0,
    dms0_request1,
    dms0_request2,
    dms0_log
  ]
}

log dms0_log {}

http_request dms0_request0 {
  method = "POST"
  url = "https://webhook.site/${env.WH_ID}"
  encoding = "json"
  body = {
    "test": "${msg.event}"
  }
}

http_request dms0_request1 {
  method = "POST"
  url = "https://webhook.site/${env.WH_ID}"
  basic_auth = {
    user: "user"
    password: "password"
  }
  encoding = "url"
  body = {
    "test": "${msg.event}"
  }
}

http_request dms0_request2 {
  method = "POST"
  url = "https://webhook.site/${env.WH_ID}"
  encoding = "raw"
  body = "${msg.event}"
}
