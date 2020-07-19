package blocks

import (
	"log"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/json"
)

type Log struct {
	SingleChannelBlock
	Text hcl.Expression `hcl:"text"`
}

func (l *Log) Start(ctx *bctx.BCtx) error {
	ch0 := l.Ch0(ctx)
	go func() {
		for msg := range ch0 {
			val := msg.Value()

			if !l.Text.Range().Empty() {
				evCtx := ctx.DefaultEvaluationContext(&msg)
				var err error
				val, err = EvaluateExpression(l.Text, evCtx)
				if err != nil {
					ctx.WriteError(err)
					msg.ReplyWithError()
					return
				}
			}

			marshal, err := json.Marshal(val, val.Type())
			if err != nil {
				log.Println(err)
			}
			log.Println(string(marshal))
			msg.Close()
		}
	}()
	return nil
}
