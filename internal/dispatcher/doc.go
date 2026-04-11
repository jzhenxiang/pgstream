// Package dispatcher provides table-level event routing for pgstream.
//
// A Dispatcher holds a set of named routes, each mapping a fully-qualified
// table name ("schema.table") to one or more sinks. When an event arrives,
// the dispatcher looks up the event's table in its routing table and forwards
// the event to every matching sink. Events that do not match any route are
// sent to the configured default sinks, if any.
//
// Example:
//
//	routes := []dispatcher.Route{
//		{Table: "public.orders", Sinks: []sink.Sink{kafkaSink}},
//		{Table: "public.users",  Sinks: []sink.Sink{webhookSink}},
//	}
//	d, err := dispatcher.New(routes, nil)
//	if err != nil { ... }
//	err = d.Dispatch(ctx, event)
package dispatcher
