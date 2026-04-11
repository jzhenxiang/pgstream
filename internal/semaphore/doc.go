// Package semaphore provides a lightweight counting semaphore for
// controlling concurrent access to shared resources in the pgstream
// pipeline.
//
// Usage:
//
//	sem, err := semaphore.New(10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Block until a slot is available.
//	if err := sem.Acquire(ctx); err != nil {
//	    return err
//	}
//	defer sem.Release()
//
// The semaphore is safe for concurrent use.
package semaphore
