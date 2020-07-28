package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
)

type Map struct {
	SingleChannelBlock
	Expr   hcl.Expression      `hcl:"expr"`
	SendTo bctx.ChannelPointer `hcl:"send_to"`
}

func (m *Map) Start(env *bctx.BEnv) error {
	sendTo := m.SendTo.SendCh(env)

	env.StartProcessing(m.Ch0(env), func(msg comm.Msg) error {
		// Executing the expression
		exprValue, err := bctx.EvaluateExpression(m.Expr, env.DefaultEvaluationContext(&msg))
		if err != nil {
			return err
		}

		// Preparing and forwarding the modified message
		ctx := msg.Ctx
		newMsg, ch := comm.NewMessageC(ctx, exprValue)
		sendTo <- newMsg

		select {
		case <-ctx.Done():
			msg.Close()
		case reply, ok := <-ch:
			if ok {
				msg.Reply(reply)
			} else {
				msg.Close()
			}
		}

		return nil
	})

	return nil
}
