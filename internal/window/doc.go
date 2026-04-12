// Package window implements a sliding time-window counter.
//
// Events are bucketed by a configurable granularity (e.g. 1 s) and
// automatically evicted once they fall outside the window duration.
//
// Typical usage:
//
//	w, _ := window.New(time.Minute, time.Second)
//	w.Add(1)
//	fmt.Println(w.Count()) // events in the last 60 s
package window
