package blocks

type BlockFactory func() Block

func BlockRegistry() map[string]BlockFactory {
	return map[string]BlockFactory{
		"http_server":      func() Block { return &HttpServer{} },
		"dead_mans_switch": func() Block { return &DeadMansSwitch{} },
		"http_request":     func() Block { return &HttpRequest{} },

		"splitter": func() Block { return &Splitter{} },
		"mux":      func() Block { return &Mux{} },

		"log": func() Block { return &Log{} },
	}
}
