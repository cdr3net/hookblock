package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
)

type Splitter struct {
	SingleChannelBlock
	Expr             hcl.Expression      `hcl:"expr"`
	SendTo           bctx.ChannelPointer `hcl:"send_to"`
	TerminateOnError bool                `hcl:"terminate_on_error,optional"`
}

func (s *Splitter) Start(env *bctx.BEnv) error {
	sendTo := s.SendTo.SendCh(env)
	env.StartProcessing(s.Ch0(env), func(msg comm.Msg) error {
		val, err := bctx.EvaluateExpression(s.Expr, env.DefaultEvaluationContext(&msg))
		if err != nil {
			return err
		}

		vals := val.AsValueSlice()
		var reqs []sendRequest
		for _, v := range vals {
			reqs = append(reqs, sendRequest{
				ctx:    msg.Ctx,
				sendTo: sendTo,
				value:  v,
			})
		}

		result, _ := sendAll(s.TerminateOnError, reqs)

		msg.Reply(result)

		return nil
	})

	return nil
}
