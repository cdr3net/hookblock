http_server server0 {
  address = "127.0.0.1:2345"
  timeout = "10s"

  endpoint {
    methods = ["POST"]
    path = "/test/{test_var}"
    send_to = test_split
  }

  endpoint {
    path = "/reset_dms/${env.DMS_SECRET}"
    send_to = reset_message_mux
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

mux reset_message_mux {
  send_to = [
    dms0,
    reset_message_log
  ]
}

log reset_message_log {
  text = msg
}

timer dms0 {
  initial_timeout = "5s"
  timeout = msg.body
  on_timeout = dms0_map
  on_repeat = dms0_map
  on_reset = dms0_map
}

map dms0_map {
  expr = {
    "event": "The event is ${msg.event}."
  }
  send_to = dms0_mux
}

mux dms0_mux {
  send_to = [
    dms0_request0,
    dms0_request1,
    dms0_request2,
    dms0_log
  ]
}

splitter test_split {
  expr = msg.body
  send_to = test_mux
}

mux test_mux {
  send_to = [
    test_log1,
    test_log2,
  ]
}

log test_log1 {
  text = msg
}

log test_log2 {
  text = msg
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
