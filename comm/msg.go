package comm

import (
	"context"
	"sync/atomic"

	"github.com/dbolotin/deadmanswitch/ctyutil"
	"github.com/zclconf/go-cty/cty"
)

type Msg struct {
	Ctx          context.Context
	replyTo      chan<- cty.Value
	valueFactory func() cty.Value
	value        *cty.Value
	answered     *int32
}

func NewMessage(ctx context.Context, replyChannel chan cty.Value, value cty.Value) Msg {
	var answered int32 = 0
	return Msg{
		Ctx:      ctx,
		replyTo:  replyChannel,
		value:    &value,
		answered: &answered,
	}
}

func NewLazyMessage(ctx context.Context, replyChannel chan cty.Value, valueFactory func() cty.Value) Msg {
	var answered int32 = 0
	return Msg{
		Ctx:          ctx,
		replyTo:      replyChannel,
		valueFactory: valueFactory,
		answered:     &answered,
	}
}

func NewMessageNoC(ctx context.Context, value cty.Value) Msg {
	return NewMessage(ctx, nil, value)
}

func NewLazyMessageNoC(ctx context.Context, valueFactory func() cty.Value) Msg {
	return NewLazyMessage(ctx, nil, valueFactory)
}

func NewMessageC(ctx context.Context, value cty.Value) (Msg, chan cty.Value) {
	replyChannel := make(chan cty.Value, 1)
	return NewMessage(ctx, replyChannel, value), replyChannel
}

func NewLazyMessageC(ctx context.Context, valueFactory func() cty.Value) (Msg, chan cty.Value) {
	replyChannel := make(chan cty.Value, 1)
	return NewLazyMessage(ctx, replyChannel, valueFactory), replyChannel
}

func (m *Msg) Value() cty.Value {
	if m.value == nil {
		val := m.valueFactory()
		m.value = &val
	}
	return *m.value
}

func (m *Msg) Reply(val cty.Value) {
	if m.replyTo == nil {
		return
	}
	if atomic.CompareAndSwapInt32(m.answered, 0, 1) {
		m.replyTo <- val
		close(m.replyTo)
	}
}

var ErrorReply = cty.ObjectVal(map[string]cty.Value{"err": ctyutil.StrNullVal})

func IsErrorReply(rep cty.Value) bool {
	return rep == ErrorReply
}

func (m *Msg) ReplyWithError() {
	m.Reply(ErrorReply)
}

func (m *Msg) Close() {
	if m.replyTo == nil {
		return
	}
	if atomic.CompareAndSwapInt32(m.answered, 0, 1) {
		close(m.replyTo)
	}
}
