package blocks

type BlockFactory func() Block

func BlockRegistry() map[string]BlockFactory {
	return map[string]BlockFactory{
		"http_server":  func() Block { return &HttpServer{} },
		"http_request": func() Block { return &HttpRequest{} },

		"timer": func() Block { return &Timer{} },

		"map":      func() Block { return &Map{} },
		"splitter": func() Block { return &Splitter{} },
		"mux":      func() Block { return &Mux{} },

		"deduplicate": func() Block { return &Deduplicate{} },

		"log": func() Block { return &Log{} },
	}
}
