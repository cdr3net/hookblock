package blocks

import (
	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/ctyutil"
)

type Deduplicate struct {
	SingleChannelBlock
	// Expr   hcl.Expression      `hcl:"expr"`
	SendTo bctx.ChannelPointer `hcl:"send_to"`
}

func (d *Deduplicate) Start(env *bctx.BEnv) error {
	sendTo := d.SendTo.SendCh(env)
	ch0 := d.Ch0(env)

	go func() {
		var previous = ctyutil.StrNullVal
		for msg := range ch0 {
			current := msg.Value()
			if previous.RawEquals(current) {
				msg.Close()
				continue
			} else {
				sendTo <- msg
				previous = current
			}
		}
	}()

	return nil
}
