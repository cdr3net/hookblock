package blocks

import "github.com/dbolotin/deadmanswitch/comm"

type Block interface {
	SetInputChannel(ch <-chan comm.Msg)
	Start()
}

type ABlock struct {
	InputChannel <-chan comm.Msg
}

func (b *ABlock) SetInputChannel(ch <-chan comm.Msg) {
	b.InputChannel = ch
}
