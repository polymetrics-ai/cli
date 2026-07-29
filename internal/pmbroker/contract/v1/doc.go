// Package contractv1 pins the CLI-side PM Broker /v1 HTTP/JSON contract
// fixtures and typed client foundations.
//
// The package is intentionally narrow: callers get typed fixture values, a
// deterministic in-memory fake broker, and a typed HTTP/JSON client foundation
// for HTTPS endpoints plus the narrow loopback/PM Broker container HTTP
// allowlist enforced by NewHTTPClient. HTTP responses and collection pages stay
// bounded. The package exposes explicit internal auth and correlation seams, but
// it does not expose authentication registry stability, provider SDKs, raw
// credentials, public gRPC, divergent socket semantics, generic HTTP requests,
// arbitrary headers, SQL, shell, runtime plugins, or live deployment resources.
// Future CLI work should consume the typed client methods here and keep
// profile/context validation decisions in later, explicitly reviewed slices.
package contractv1
