// Package window implements a sliding time-window counter.
//
// Events are bucketed by a configurable granularity (e.g. 1 s) and
// automatically evicted once they fall outside the window duration.
// The window duration must be a positive multiple of the bucket granularity.
//
// Typical usage:
//
//	w, err := window.New(time.Minute, time.Second)
//	if err != nil {
//		log.Fatal(err)
//	}
//	w.Add(1)
//	fmt.Println(w.Count()) // events in the last 60 s
//
// Thread safety: all methods on Window are safe for concurrent use.
package window
