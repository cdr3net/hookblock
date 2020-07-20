package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/dbolotin/deadmanswitch/ctyutil"
	"github.com/zclconf/go-cty/cty"
)

type Block interface {
	GetValue(env *bctx.BEnv) cty.Value
	Start(env *bctx.BEnv) error
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

func (b *IsolatedBlock) GetValue(ctx *bctx.BEnv) cty.Value {
	return ctyutil.StrNullVal
}

type SingleChannelBlock struct {
	ABlock
	ICh0 *bctx.ChannelPointer
}

func (b *SingleChannelBlock) GetValue(env *bctx.BEnv) cty.Value {
	if b.ICh0 == nil {
		b.ICh0 = env.NewChannel()
	}
	return b.ICh0.ToCty()
}

func (b *SingleChannelBlock) Ch0(ctx *bctx.BEnv) <-chan comm.Msg {
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
