package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
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

type IsolatedBlock struct {
	ABlock
}

func (b *IsolatedBlock) GetValue(ctx *bctx.BCtx) cty.Value {
	return cty.NilVal
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

func StrOrDefault(str *string, def string) string {
	if str == nil {
		return def
	} else {
		return *str
	}
}

func IntOrDefault(str *int, def int) int {
	if str == nil {
		return def
	} else {
		return *str
	}
}

func UInt64OrDefault(str *uint64, def uint64) uint64 {
	if str == nil {
		return def
	} else {
		return *str
	}
}

func UInt32OrDefault(str *uint32, def uint32) uint32 {
	if str == nil {
		return def
	} else {
		return *str
	}
}

func Int64OrDefault(str *int64, def int64) int64 {
	if str == nil {
		return def
	} else {
		return *str
	}
}

func Int32OrDefault(str *int32, def int32) int32 {
	if str == nil {
		return def
	} else {
		return *str
	}
}

func EvaluateExpression(expr hcl.Expression, ctx *hcl.EvalContext) (cty.Value, error) {
	value, diag := expr.Value(ctx)
	if diag.HasErrors() || !value.IsWhollyKnown() {
		return value, diag
	}
	return value, nil
}
