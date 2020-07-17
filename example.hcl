http_server "server0" {
  address = "127.0.0.1:2345"

  endpoint "GET" {
    path = "/ddds"
    send_to = dms0
  }

  endpoint "GET" {
    path = "/asdfffs"
    send_to = dms0
  }
}
