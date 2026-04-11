// Package router assembles the HTTP layer of pgstream.
//
// It composes the health-check handler and optional middleware (request
// logging, panic recovery, HMAC signature verification) into a single
// http.Handler that can be passed directly to an http.Server.
//
// Usage:
//
//	h, err := router.New(router.Config{
//		SigningSecret: os.Getenv("PGSTREAM_SIGNING_SECRET"),
//		Logger:        logger,
//	}, hc)
//	if err != nil {
//		log.Fatal(err)
//	}
//	http.ListenAndServe(":8080", h)
package router
