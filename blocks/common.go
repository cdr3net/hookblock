package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/zclconf/go-cty/cty"
)

type Block interface {
	GetValue(ctx *bctx.BCtx) cty.Value
	Start(ctx *bctx.BCtx) error
	SetId(id string)
	GetId() string
}

type ABlock struct {
	Id string
}

func (b *ABlock) SetId(id string) {
	b.Id = id
}

func (b *ABlock) GetId() string {
	return b.Id
}

type SingleChannelBlock struct {
	ABlock
	ICh0 *bctx.ChannelPointer
}

func (b *SingleChannelBlock) GetValue(ctx *bctx.BCtx) cty.Value {
	if b.ICh0 == nil {
		b.ICh0 = ctx.NewChannel()
	}
	return b.ICh0.ToCty()
}

func (b *SingleChannelBlock) Ch0(ctx *bctx.BCtx) <-chan comm.Msg {
	return b.ICh0.RecvCh(ctx)
}
