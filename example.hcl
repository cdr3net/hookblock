http_server server0 {
  address = "127.0.0.1:2345"
  timeout = "10s"

  endpoint {
    methods = ["POST"]
    path = "/ddds"
    send_to = test0
  }

  endpoint {
    methods = ["GET"]
    path = "/asdfffs"
    send_to = dms0
  }
}

log log0 {}

dead_mans_switch dms0 {
  timeout = "2m"
  send_to = [test0]
}

http_request test0 {
  method = "POST"
  url = "https://webhook.site/5959f754-dac5-4b3f-86da-731e645d1926"
  body = {
    "guga": "${msg.body.a}"
  }
}