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

func CreateMessage(ctx context.Context, replyChannel chan cty.Value, value cty.Value) Msg {
	return Msg{
		Ctx:     ctx,
		ReplyTo: replyChannel,
		value:   &value,
	}
}

func CreateLazyMessage(ctx context.Context, replyChannel chan cty.Value, valueFactory func() cty.Value) Msg {
	return Msg{
		Ctx:          ctx,
		ReplyTo:      replyChannel,
		valueFactory: valueFactory,
	}
}

func (m *Msg) Value() cty.Value {
	if m.value == nil {
		val := m.valueFactory()
		m.value = &val
	}
	return *m.value
}
