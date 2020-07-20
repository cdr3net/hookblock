package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
)

type Mux struct {
	SingleChannelBlock
	TerminateOnError bool                  `hcl:"terminate_on_error,optional"`
	SendTo           []bctx.ChannelPointer `hcl:"send_to"`
}

func (s *Mux) Start(env *bctx.BEnv) error {
	var sendTo []chan<- comm.Msg
	for _, s := range s.SendTo {
		sendTo = append(sendTo, s.SendCh(env))
	}

	env.StartProcessing(s.Ch0(env), func(msg comm.Msg) error {
		var reqs []sendRequest
		for _, s := range sendTo {
			reqs = append(reqs, sendRequest{
				ctx:    msg.Ctx,
				sendTo: s,
				value:  msg.Value(),
			})
		}

		result, _ := sendAll(s.TerminateOnError, reqs)

		msg.Reply(result)

		return nil
	})

	return nil
}
