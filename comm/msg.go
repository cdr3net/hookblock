package comm

import (
	"context"

	"github.com/dbolotin/deadmanswitch/ctyutil"
	"github.com/zclconf/go-cty/cty"
)

type Msg struct {
	Ctx          context.Context
	replyTo      chan<- cty.Value
	valueFactory func() cty.Value
	value        *cty.Value
}

func NewMessage(ctx context.Context, replyChannel chan cty.Value, value cty.Value) Msg {
	return Msg{
		Ctx:     ctx,
		replyTo: replyChannel,
		value:   &value,
	}
}

func NewLazyMessage(ctx context.Context, replyChannel chan cty.Value, valueFactory func() cty.Value) Msg {
	return Msg{
		Ctx:          ctx,
		replyTo:      replyChannel,
		valueFactory: valueFactory,
	}
}

func NewMessageNoC(ctx context.Context, value cty.Value) Msg {
	return NewMessage(ctx, nil, value)
}

func NewLazyMessageNoC(ctx context.Context, valueFactory func() cty.Value) Msg {
	return NewLazyMessage(ctx, nil, valueFactory)
}

func NewMessageC(ctx context.Context, value cty.Value) (Msg, chan cty.Value) {
	replyChannel := make(chan cty.Value)
	return NewMessage(ctx, replyChannel, value), replyChannel
}

func NewLazyMessageC(ctx context.Context, valueFactory func() cty.Value) (Msg, chan cty.Value) {
	replyChannel := make(chan cty.Value)
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
	m.replyTo <- val
	close(m.replyTo)
}

func (m *Msg) ReplyWithError() {
	m.Reply(cty.ObjectVal(map[string]cty.Value{"err": ctyutil.StrNullVal}))
}

func (m *Msg) Close() {
	if m.replyTo == nil {
		return
	}
	close(m.replyTo)
}
