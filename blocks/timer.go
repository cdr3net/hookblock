package blocks

import (
	"context"
	"errors"
	"time"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/comm"
	"github.com/hashicorp/hcl/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/zclconf/go-cty/cty"
)

type Timer struct {
	SingleChannelBlock
	InitialTimeout *string              `hcl:"initial_timeout,optional"`
	Timeout        hcl.Expression       `hcl:"timeout"`
	RepeatAfter    *string              `hcl:"repeat_after,optional"`
	BackoffFactor  float64              `hcl:"backoff_factor,optional"`
	OnTimeout      *bctx.ChannelPointer `hcl:"on_timeout,optional"`
	OnReset        *bctx.ChannelPointer `hcl:"on_reset,optional"`
	OnRepeat       *bctx.ChannelPointer `hcl:"on_repeat,optional"`
}

var (
	timerResetsVec   = promauto.NewCounterVec(prometheus.CounterOpts{Name: "timer_reset"}, []string{"block"})
	timerTimeoutsVec = promauto.NewCounterVec(prometheus.CounterOpts{Name: "timer_timeout"}, []string{"block"})
	timerRepeatsVec  = promauto.NewCounterVec(prometheus.CounterOpts{Name: "timer_repeats"}, []string{"block"})
)

const ZeroDuration = 0 * time.Minute

func (t *Timer) Start(env *bctx.BEnv) error {
	var err error

	initialTimeout := ZeroDuration
	if t.InitialTimeout != nil {
		initialTimeout, err = time.ParseDuration(*t.InitialTimeout)
		if err != nil {
			return err
		}
	}

	repeatAfter := ZeroDuration
	if t.RepeatAfter != nil {
		repeatAfter, err = time.ParseDuration(*t.RepeatAfter)
		if err != nil {
			return err
		}
	}

	// BackoffFactor must be greater then one
	if t.BackoffFactor < 1 {
		t.BackoffFactor = 1
	}

	// Monitoring
	pLabels := prometheus.Labels{
		"block": t.Id,
	}
	mResets := timerResetsVec.With(pLabels)
	mTimeouts := timerTimeoutsVec.With(pLabels)
	mRepeats := timerRepeatsVec.With(pLabels)

	var onResetCh, onTimeoutCh, onRepeatCh chan<- comm.Msg

	if t.OnReset != nil {
		onResetCh = t.OnReset.SendCh(env)
	}
	if t.OnTimeout != nil {
		onTimeoutCh = t.OnTimeout.SendCh(env)
	}
	if t.OnRepeat != nil {
		onRepeatCh = t.OnRepeat.SendCh(env)
	}

	ch0 := t.Ch0(env)
	go func() {
		var timer *time.Timer = nil
		var timerChannel <-chan time.Time = nil
		var cancelCurrentRequest context.CancelFunc = nil

		if initialTimeout != ZeroDuration {
			timer = time.NewTimer(initialTimeout)
			timerChannel = timer.C
			cancelCurrentRequest = sendTimerMsg("reset", initialTimeout, onResetCh)
		}

		currentTimeout := initialTimeout
		isOnRepeat := false

		for {
			select {
			case msg := <-ch0:
				mResets.Inc()
				isOnRepeat = false

				// Executing timeout expression
				timeoutValue, err := bctx.EvaluateExpression(t.Timeout, env.DefaultEvaluationContext(&msg))
				if err != nil {
					env.WriteError(err)
					msg.ReplyWithError()
					continue
				}

				if timeoutValue.Type() == cty.String {
					currentTimeout, err = time.ParseDuration(timeoutValue.AsString())
					if err != nil {
						env.WriteError(err)
						msg.ReplyWithError()
						continue
					}
				} else if timeoutValue.Type() == cty.Number {
					val, _ := timeoutValue.AsBigFloat().Int64()
					currentTimeout = time.Duration(val) * time.Second
				} else {
					if err != nil {
						env.WriteError(errors.New("Wrong timeout type: " + timeoutValue.Type().GoString()))
						msg.ReplyWithError()
						continue
					}
				}

				// Cancelling context of previously sent downstream requests
				if cancelCurrentRequest != nil {
					cancelCurrentRequest()
					cancelCurrentRequest = nil
				}

				// Resetting timer
				if timer != nil {
					timer.Stop()
					timer = nil
					timerChannel = nil
				}

				if currentTimeout != ZeroDuration {
					timer = time.NewTimer(currentTimeout)
					timerChannel = timer.C
					cancelCurrentRequest = sendTimerMsg("reset", currentTimeout, onResetCh)
				}

				// Reporting to the upstream block that we processed the message
				msg.Close()

			case <-timerChannel:
				var event string
				var targetCh chan<- comm.Msg
				if isOnRepeat {
					mRepeats.Inc()
					event = "repeat"
					currentTimeout = time.Duration(t.BackoffFactor * float64(currentTimeout))
					targetCh = onRepeatCh
				} else {
					mTimeouts.Inc()
					event = "timeout"
					currentTimeout = repeatAfter
					targetCh = onTimeoutCh
				}

				// Cancelling context of previously sent downstream request
				if cancelCurrentRequest != nil {
					cancelCurrentRequest()
					cancelCurrentRequest = nil
				}

				if timer != nil {
					timer.Stop()
					timer = nil
					timerChannel = nil
				}

				isOnRepeat = true
				if currentTimeout != ZeroDuration {
					timer = time.NewTimer(currentTimeout)
					timerChannel = timer.C
				}

				cancelCurrentRequest = sendTimerMsg(event, currentTimeout, targetCh)
			}
		}
	}()

	return nil
}

func sendTimerMsg(event string, timeout time.Duration, to chan<- comm.Msg) context.CancelFunc {
	if to == nil {
		return nil
	}

	cCtx, cancelCurrentRequest := context.WithCancel(context.Background())
	to <- comm.NewMessageNoC(
		cCtx,
		cty.ObjectVal(map[string]cty.Value{
			"event":   cty.StringVal(event),
			"timeout": cty.NumberFloatVal(float64(timeout) / float64(time.Second)),
		}),
	)

	return cancelCurrentRequest
}
