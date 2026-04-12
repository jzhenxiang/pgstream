// Package heartbeat keeps Postgres replication slots alive by sending periodic
// standby status updates (heartbeats) over the replication connection.
//
// Without regular feedback Postgres may reclaim WAL segments that pgstream still
// needs, causing the replication slot to become invalidated. The Heartbeat type
// runs in its own goroutine and calls Sender.SendStandbyStatus on a configurable
// interval (default 10 s).
//
// # Background
//
// PostgreSQL requires that a standby (or logical replication client) send
// periodic status updates so the primary knows the client is still alive and
// which WAL positions have been safely processed. If no update is received
// within wal_sender_timeout the connection is dropped and the replication slot
// may be invalidated, potentially causing data loss on restart.
//
// # Usage
//
//	hb, err := heartbeat.New(myReplicationConn, heartbeat.Config{
//		Interval: 5 * time.Second,
//	})
//	if err != nil { … }
//	go hb.Run(ctx)
//
// Run blocks until ctx is cancelled. It is safe to call Run from a goroutine
// and cancel the context to perform a clean shutdown.
package heartbeat
