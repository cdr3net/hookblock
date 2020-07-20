package bctx

import (
	"fmt"
	"log"
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

func (c ChannelPointer) SendCh(env *BEnv) chan<- comm.Msg {
	return env.channel(c.Id)
}

func (c ChannelPointer) RecvCh(env *BEnv) <-chan comm.Msg {
	return env.channel(c.Id)
}

type BEnv struct {
	DefaultVariables map[string]cty.Value
	dw               hcl.DiagnosticWriter
	channels         map[string]chan comm.Msg
	i                uint64
}

func NewCtx(dw hcl.DiagnosticWriter) *BEnv {
	return &BEnv{
		DefaultVariables: make(map[string]cty.Value),
		dw:               dw,
		channels:         make(map[string]chan comm.Msg),
	}
}

func (ctx *BEnv) DefaultEvaluationContext(msg *comm.Msg) *hcl.EvalContext {
	// Creating the evaluation context
	evCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"msg": msg.Value(),
		},
	}

	// Adding default variables
	for k, v := range ctx.DefaultVariables {
		evCtx.Variables[k] = v
	}

	return evCtx
}

func (ctx *BEnv) StartProcessing(msgCh <-chan comm.Msg, handler func(msg comm.Msg) error) {
	go func() {
		for m := range msgCh {
			// Saving msg to a separate variable to use it in a forked goroutine
			// Important: "m" must not be used
			msg := m

			// Each request processed in a separate goroutine
			go func() {
				// Handling panics
				defer func() {
					if r := recover(); r != nil {
						ctx.WriteError(fmt.Errorf("%s", r))
						msg.ReplyWithError()
					}
				}()

				// Executing main handler
				err := handler(msg)

				if err != nil {
					// Reply with error
					ctx.WriteError(err)
					msg.ReplyWithError()
				} else {
					// Ensure message reply channel is closed
					msg.Close()
				}
			}()
		}

		// Should never reach this statement
		log.Fatalln("Communication channel closed.")
	}()
}

func EvaluateExpression(expr hcl.Expression, ctx *hcl.EvalContext) (cty.Value, error) {
	value, diag := expr.Value(ctx)
	if diag.HasErrors() || !value.IsWhollyKnown() {
		return value, diag
	}
	return value, nil
}

func (ctx *BEnv) nextChId() string {
	ctx.i++
	return "ch" + strconv.FormatUint(ctx.i, 10)
}

func (ctx *BEnv) channel(id string) chan comm.Msg {
	channel, ok := ctx.channels[id]
	if !ok {
		panic("Communication channel with such id not registered: " + id)
	} else {
		return channel
	}
}

// Registers channel and returns its id
func (ctx *BEnv) NewChannel() *ChannelPointer {
	ch := make(chan comm.Msg, 1) // 1 -> Safer in terms of stupid deadlocks
	id := ctx.nextChId()
	ctx.channels[id] = ch
	return &ChannelPointer{Id: id}
}

func (ctx *BEnv) WriteError(err error) {
	// TODO synchronize, limit rate, monitoring
	if diagnostic, ok := err.(*hcl.Diagnostic); ok {
		ctx.WriteDiagnostic(diagnostic)
	} else {
		log.Println(err)
	}
}

func (ctx *BEnv) WriteDiagnostic(diagnostic *hcl.Diagnostic) {
	// TODO synchronize, limit rate, monitoring
	err := ctx.dw.WriteDiagnostic(diagnostic)
	if err != nil {
		log.Fatalln(err)
	}
}

func (ctx *BEnv) WriteDiagnostics(diagnostics hcl.Diagnostics) {
	// TODO synchronize, limit rate, monitoring
	err := ctx.dw.WriteDiagnostics(diagnostics)
	if err != nil {
		log.Fatalln(err)
	}
}
