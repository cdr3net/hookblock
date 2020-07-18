package blocks

import (
	"log"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/zclconf/go-cty/cty/json"
)

type Log struct {
	SingleChannelBlock
}

func (l *Log) Start(ctx *bctx.BCtx) error {
	ch0 := l.Ch0(ctx)
	go func() {
		for msg := range ch0 {
			val := msg.Value()
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
