// Package contractv1 pins the synthetic PM Broker /v1 HTTP/JSON contract
// fixtures accepted by PM Broker PR #35 for CLI-side tests.
//
// The package is intentionally narrow: callers get typed fixture values and a
// deterministic in-memory fake broker client for profile, context, and
// execution-plan lanes. It does not expose production transport setup,
// authentication registry stability, provider SDKs, raw credentials, generic
// HTTP, arbitrary headers, SQL, shell, runtime plugins, or live deployment
// resources. Future CLI work should consume the typed client methods here and
// keep production broker transport/auth decisions in later, explicitly reviewed
// slices.
package contractv1
