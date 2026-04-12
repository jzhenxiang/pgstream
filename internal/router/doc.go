// Package router builds and wires the HTTP router used by pgstream's webhook
// receiver. It exposes:
//
//   - New: constructs the mux with health-check and event routes.
//   - Chain: composes middleware in left-to-right (outermost-first) order.
//   - WithTimeout: attaches a per-request deadline to the context.
//   - WithSigning: verifies HMAC-SHA256 request signatures produced by the
//     middleware.Signer; a blank secret disables verification.
package router
