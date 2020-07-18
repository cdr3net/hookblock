package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
)

type DeadMansSwitch struct {
	SingleChannelBlock
	Timeout       string  `hcl:"timeout"`
	RepeatAfter   string  `hcl:"repeat_after,optional"`
	BackoffFactor float64 `hcl:"backoff_factor,optional"`

	SendTo []bctx.ChannelPointer `hcl:"send_to"`
}

func (d *DeadMansSwitch) Start(ctx *bctx.BCtx) error {
	// timeout, err := time.ParseDuration(d.Timeout)
	// if err != nil {
	// 	return err
	// }
	// repeatAfter, err := time.ParseDuration(d.RepeatAfter)
	// if err != nil {
	// 	return err
	// }
	//
	// ch0 := d.Ch0(ctx)
	//
	// go func() {
	//
	// 	timer := time.NewTimer(timeout)
	// 	lastTimeout := timeout
	//
	// 	onRepeat := false
	//
	// 	for {
	// 		select {
	// 		case m := <-ch0:
	// 			resetCounter.Inc()
	// 			onRepeat = false
	// 			lastTimeout = endpoint.Timeout
	// 			timer.Reset(lastTimeout)
	//
	// 		case <-timer.C:
	// 			var event string
	// 			if onRepeat {
	// 				repeatCounter.Inc()
	// 				event = "Repeat"
	// 				lastTimeout = time.Duration(e.BackoffFactor * float64(lastTimeout))
	// 			} else {
	// 				timeoutCounter.Inc()
	// 				event = "Timeout"
	// 				lastTimeout = e.RepeatAfter
	// 			}
	//
	// 			onRepeat = true
	// 			timer.Reset(lastTimeout)
	// 		}
	// 	}
	//
	// 	for m := range h.ICh0.RecvCh(ctx) {
	// 		fmt.Println("asd", m.Value())
	// 		// close(m.ReplyTo)
	// 	}
	// }()

	return nil
}
