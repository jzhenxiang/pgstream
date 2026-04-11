// Package heartbeat keeps Postgres replication slots alive by sending periodic
// standby status updates (heartbeats) over the replication connection.
//
// Without regular feedback Postgres may reclaim WAL segments that pgstream still
// needs, causing the replication slot to become invalidated. The Heartbeat type
// runs in its own goroutine and calls Sender.SendStandbyStatus on a configurable
// interval (default 10 s).
//
// Usage:
//
//	hb, err := heartbeat.New(myReplicationConn, heartbeat.Config{
//		Interval: 5 * time.Second,
//	})
//	if err != nil { … }
//	go hb.Run(ctx)
package heartbeat
