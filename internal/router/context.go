package router

import (
	"context"
	"time"
)

// withDeadline is a thin wrapper so tests can swap it out.
var withDeadline = func(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}
