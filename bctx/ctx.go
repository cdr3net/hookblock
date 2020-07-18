package bctx

import (
	"strconv"

	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ChannelPointer struct {
	Id string `cty:"id"`
}

func (c ChannelPointer) ToCty() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{"id": cty.StringVal(c.Id)})
}

func (c ChannelPointer) SendCh(ctx *BCtx) chan<- comm.Msg {
	return ctx.channel(c.Id)
}

func (c ChannelPointer) RecvCh(ctx *BCtx) <-chan comm.Msg {
	return ctx.channel(c.Id)
}

type BCtx struct {
	dw       hcl.DiagnosticWriter
	channels map[string]chan comm.Msg
	i        uint64
}

func NewCtx(dw hcl.DiagnosticWriter) *BCtx {
	return &BCtx{
		dw:       dw,
		channels: make(map[string]chan comm.Msg),
	}
}

func (ctx *BCtx) nextChId() string {
	ctx.i++
	return "ch" + strconv.FormatUint(ctx.i, 10)
}

func (ctx *BCtx) channel(id string) chan comm.Msg {
	channel, ok := ctx.channels[id]
	if !ok {
		panic("Communication channel with such id not registered: " + id)
	} else {
		return channel
	}
}

// Registers channel and returns its id
func (ctx *BCtx) NewChannel() *ChannelPointer {
	ch := make(chan comm.Msg)
	id := ctx.nextChId()
	ctx.channels[id] = ch
	return &ChannelPointer{Id: id}
}

func (ctx *BCtx) WriteDiagnostic(diagnostic *hcl.Diagnostic) {
	// TODO synchronize, limit rate
	ctx.dw.WriteDiagnostic(diagnostic)
}

func (ctx *BCtx) WriteDiagnostics(diagnostics hcl.Diagnostics) {
	// TODO synchronize, limit rate
	ctx.dw.WriteDiagnostics(diagnostics)
}
