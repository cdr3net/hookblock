package blocks

import "github.com/dbolotin/deadmanswitch/comm"

type Endpoint struct {
	Method string        `hcl:"method,label"`
	Path   string        `hcl:"path"`
	SendTo chan comm.Msg `hcl:"send_to"`
}

type HttpServer struct {
	ABlock
	Address   string     `hcl:"address"`
	Endpoints []Endpoint `hcl:"endpoint,block"`
}
