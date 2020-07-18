package comm

import (
	"context"

	"github.com/zclconf/go-cty/cty"
)

type Msg struct {
	Ctx          context.Context
	ReplyTo      chan<- cty.Value
	valueFactory func() cty.Value
	value        *cty.Value
}

func NewMessage(ctx context.Context, replyChannel chan cty.Value, value cty.Value) Msg {
	return Msg{
		Ctx:     ctx,
		ReplyTo: replyChannel,
		value:   &value,
	}
}

func NewLazyMessage(ctx context.Context, replyChannel chan cty.Value, valueFactory func() cty.Value) Msg {
	return Msg{
		Ctx:          ctx,
		ReplyTo:      replyChannel,
		valueFactory: valueFactory,
	}
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
	m.ReplyTo <- val
	close(m.ReplyTo)
}

func (m *Msg) ReplyWithError() {
	m.Reply(cty.ObjectVal(map[string]cty.Value{"err": cty.NilVal}))
}

func (m *Msg) Close() {
	close(m.ReplyTo)
}
