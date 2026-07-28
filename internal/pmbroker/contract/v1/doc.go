// Package contractv1 pins the synthetic PM Broker /v1 HTTP/JSON contract
// fixtures accepted by PM Broker PR #35 for CLI-side tests.
//
// The package is intentionally narrow: callers get typed fixture values, a
// deterministic in-memory fake broker, and a typed HTTP/JSON client foundation
// for loopback or remote/container endpoints. It exposes explicit internal auth
// and correlation seams, but it does not expose authentication registry
// stability, provider SDKs, raw credentials, public gRPC, divergent socket
// semantics, generic HTTP requests, arbitrary headers, SQL, shell, runtime
// plugins, or live deployment resources. Future CLI work should consume the
// typed client methods here and keep profile/context validation decisions in
// later, explicitly reviewed slices.
package contractv1
