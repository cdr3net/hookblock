package blocks

import (
	"errors"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
)

type Split struct {
	SingleChannelBlock
	Expr   hcl.Expression      `hcl:"expr"`
	SendTo bctx.ChannelPointer `hcl:"send_to"`
}

type res struct {
	i       int
	success bool
}

func (s *Split) Start(env *bctx.BEnv) error {
	sendTo := s.SendTo.SendCh(env)
	env.StartProcessing(s.Ch0(env), func(msg comm.Msg) error {
		val, err := bctx.EvaluateExpression(s.Expr, env.DefaultEvaluationContext(&msg))
		if err != nil {
			return err
		}

		vals := val.AsValueSlice()
		count := len(vals)
		results := make(chan res, count)
		for i, v := range vals {
			ii := i
			m, ch := comm.NewMessageC(msg.Ctx, v)
			go func() {
				r, hasReply := <-ch
				results <- res{
					i:       ii,
					success: !hasReply || !comm.IsErrorReply(r),
				}
			}()
			sendTo <- m
		}

		for i := 0; i < count; i++ {
			r := <-results
			if !r.success {
				return errors.New("error from downstream in split")
			}
		}

		return nil
	})

	return nil
}
