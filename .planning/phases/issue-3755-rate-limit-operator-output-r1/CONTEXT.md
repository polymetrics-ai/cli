# Context — issue-3755-rate-limit-operator-output-r1

## Contract and inline GSD fallback

- Parent foundation: #3750. Completed prerequisites are #3874 declaration/admission and #3877 enforcement. This issue owns the final operator-visible observation layer only.
- `scripts/gsd doctor`, all five `scripts/gsd sources` lookups, and `go run ./cmd/agentcontractgen check` passed on 2026-08-06. The generated `discuss-phase` prompt is executed inline: Pi interactive execution is unavailable and the canonical single-worker contract prohibits role spawning.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, and `golang-concurrency`.
- `internal/coordination/rate_limits.go` and `internal/connectors/engine/rate_limit_runtime.go` are the enforcement seams. `connsdk.RateLimitObservation` and `RateLimitError` already deliberately exclude headers, URLs, request bodies, credentials, scope projections, and raw subjects.

## Fixed implementation decisions

1. Every reported run gets one bounded, secret-free rate-limit summary. Its declaration state is `declared`, `unknown`, `not_applicable`, or `undeclared`; an absent `rate_limits.json` is always `undeclared`, never unlimited.
2. A declared policy report contains only policy ID, declared subject kind, and a structural selection reason (`all`, endpoint, tier, and/or auth type). It never includes the subject config key's runtime value, a raw binding, opaque scope key, credentials, or `CredentialRevision`.
3. Pacing is recorded as an aggregate duration, not an unbounded request log. Provider observations expose only safe typed fields (latest reported remaining budget, reset/retry timing, and 429 count). Provider retry waits and ordinary requester latency are aggregated separately from local pacing.
4. Existing admission, policy resolution, retry, limiter, and reset decisions remain unchanged. Telemetry is observational around those seams and cannot alter the chosen policy, wait duration, request, or retry.
5. The normal ETL `Run` result is the durable summary carrier, so the existing human `pm etl run` and JSON envelope report the same data. Help/manual/website documents describe the added `run.rate_limit` result field; no new command or flag is introduced.

## Changed-path scope

- `internal/connectors/`: bounded report type and the runtime configuration reference only.
- `internal/coordination/`, `internal/connectors/connsdk/`, and `internal/connectors/engine/`: observational hooks and safe aggregation around the existing limiter/requester.
- `internal/app/` and `internal/cli/`: attach summaries to ETL runs and render them for human/JSON output.
- `internal/connectors/engine/testdata/`: declared test-only fixture bundles only; production declarations and `defs/defs.go` remain unchanged.
- `docs/cli/**`, `website/**`, and generated help/manual artifacts as required by the existing parity generator.

## Exclusions

- No policy/admission/pacing/resolution changes, dependencies, production `rate_limits.json`, credentialed connector checks, raw request logging, redaction/masking behavior, generic HTTP/SQL write surface, or use of `CredentialRevision` for any observation or identity purpose.
