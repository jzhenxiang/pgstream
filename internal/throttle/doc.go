// Package throttle implements an adaptive delay mechanism for the pgstream
// pipeline. When a downstream sink (Kafka, webhook, etc.) signals that it
// cannot keep up, the caller should invoke Increase to lengthen the inter-event
// pause. As the sink recovers, Decrease should be called to reduce the delay
// back toward zero.
//
// Typical usage:
//
//	th, _ := throttle.New(throttle.DefaultConfig())
//
//	for _, ev := range events {
//		if err := sink.Send(ev); err != nil {
//			th.Increase()
//		} else {
//			th.Decrease()
//		}
//		if err := th.Wait(ctx); err != nil {
//			return err
//		}
//	}
package throttle
