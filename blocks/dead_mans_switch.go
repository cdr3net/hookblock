package blocks

import (
	"context"
	"time"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/zclconf/go-cty/cty"
)

type DeadMansSwitch struct {
	SingleChannelBlock
	Timeout       string  `hcl:"timeout"`
	RepeatAfter   *string `hcl:"repeat_after,optional"`
	BackoffFactor float64 `hcl:"backoff_factor,optional"`

	SendTo []bctx.ChannelPointer `hcl:"send_to"`
}

var (
	dmsResetsVec   = promauto.NewCounterVec(prometheus.CounterOpts{Name: "dead_mans_switch_reset"}, []string{"block"})
	dmsTimeoutsVec = promauto.NewCounterVec(prometheus.CounterOpts{Name: "dead_mans_switch_timeout"}, []string{"block"})
	dmsRepeatsVec  = promauto.NewCounterVec(prometheus.CounterOpts{Name: "dead_mans_switch_repeats"}, []string{"block"})
)

const ZeroDuration = 0 * time.Minute

func (d *DeadMansSwitch) Start(env *bctx.BEnv) error {
	timeout, err := time.ParseDuration(d.Timeout)
	if err != nil {
		return err
	}

	repeatAfter := ZeroDuration
	if d.RepeatAfter != nil {
		repeatAfter, err = time.ParseDuration(*d.RepeatAfter)
		if err != nil {
			return err
		}
	}

	// BackoffFactor must be greater then one
	if d.BackoffFactor < 1 {
		d.BackoffFactor = 1
	}

	pLabels := prometheus.Labels{
		"block": d.Id,
	}
	mResets := dmsResetsVec.With(pLabels)
	mTimeouts := dmsTimeoutsVec.With(pLabels)
	mRepeats := dmsRepeatsVec.With(pLabels)

	var sendTo []chan<- comm.Msg
	for _, s := range d.SendTo {
		sendTo = append(sendTo, s.SendCh(env))
	}

	ch0 := d.Ch0(env)
	go func() {
		timer := time.NewTimer(timeout)
		lastTimeout := timeout
		onRepeat := false
		var cancelCurrentRequest context.CancelFunc = nil

		for {
			select {
			case m := <-ch0:
				mResets.Inc()
				onRepeat = false
				lastTimeout = timeout

				// Cancelling context of previously sent downstream requests
				if cancelCurrentRequest != nil {
					cancelCurrentRequest()
					cancelCurrentRequest = nil
				}

				timer.Reset(lastTimeout)
				m.Close()

			case <-timer.C:
				var event string
				if onRepeat {
					mRepeats.Inc()
					event = "repeat"
					lastTimeout = time.Duration(d.BackoffFactor * float64(lastTimeout))
				} else {
					mTimeouts.Inc()
					event = "timeout"
					lastTimeout = repeatAfter
				}

				// Cancelling context of previously sent downstream requests
				if cancelCurrentRequest != nil {
					cancelCurrentRequest()
					cancelCurrentRequest = nil
				}

				onRepeat = true
				if lastTimeout == ZeroDuration {
					timer.Stop()
				} else {
					timer.Reset(lastTimeout)
				}

				// Sending downstream messages
				var cCtx context.Context
				cCtx, cancelCurrentRequest = context.WithCancel(context.Background())
				for _, s := range sendTo {
					msg := comm.NewMessageNoC(cCtx, cty.ObjectVal(map[string]cty.Value{
						"event": cty.StringVal(event),
					}))
					s <- msg
				}
			}
		}
	}()

	return nil
}
