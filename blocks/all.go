package blocks

type BlockFactory func() Block

func BlockRegistry() map[string]BlockFactory {
	return map[string]BlockFactory{
		"http_server":      func() Block { return &HttpServer{} },
		"dead_mans_switch": func() Block { return &DeadMansSwitch{} },
		"http_request":     func() Block { return &HttpRequest{} },
		"split":            func() Block { return &Split{} },
		"log":              func() Block { return &Log{} },
	}
}
