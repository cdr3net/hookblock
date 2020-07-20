package blocks

import (
	"log"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/json"
)

type Log struct {
	SingleChannelBlock
	Text hcl.Expression `hcl:"text"`
}

func (l *Log) Start(env *bctx.BEnv) error {
	env.StartProcessing(l.Ch0(env), func(msg comm.Msg) error {
		val := msg.Value()

		if !l.Text.Range().Empty() {
			evCtx := env.DefaultEvaluationContext(&msg)
			var err error
			val, err = bctx.EvaluateExpression(l.Text, evCtx)
			if err != nil {
				return err
			}
		}

		marshal, err := json.Marshal(val, val.Type())
		if err != nil {
			log.Println(err)
		}
		log.Println(string(marshal))

		return nil
	})

	return nil
}
