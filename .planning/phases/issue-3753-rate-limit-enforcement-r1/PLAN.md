# Phase issue-3753-rate-limit-enforcement-r1 — declared rate-limit enforcement

## Required skills

`golang-how-to`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-testing`; `golang-context`; `golang-concurrency`; and `golang-documentation` were loaded. This is an inline GSD fallback: `discuss-phase` and `plan-phase --tdd` prompts were generated because Pi roles are unavailable and role spawning is forbidden.

## TDD slices

| Slice | RED proof | GREEN implementation | Guard |
| --- | --- | --- | --- |
| A — scoped registry (#3754) | Requests under the same opaque scope do not share capacity; cancelled waits can hang; raw keys can be retained | local registry implements fixed/sliding/token/leaky budgets and context-aware clock waiting | test linked/unlinked bindings, credential-revision independence, unsupported scope refusal, response cost tightening |
| B — policy resolver (#3753) | Declared test bundles do not select by endpoint/tier/auth or attach an admission | resolve matching policies with the declared non-secret subject config key and attach a shared registry limiter | no declaration/unknown/not-applicable remains byte-identical; policy scope is opaque |
| C — requester coverage | Check, read pages, direct reads, direct writes (JSON/form/multipart), declarative writes, and binary streams can reach a server without resolver admission | route every engine-created requester through `Runtime.RequesterFor` before send | a failing per-path counter test protects every path; strict write no-replay stays intact |
| D — model and precedence | Cost header, fixed/sliding/token/leaky capacity, and legacy `base.rate_limit` are either ignored or collapse to one RPM delay | honor each declared budget and use an observed actual point cost only to tighten state | injected clock only; legacy limiter remains independently invoked |

## Commit checkpoints

1. Commit planning/TDD artifacts.
2. Retain a red focused test checkpoint before implementation.
3. Commit the green registry/resolver/wiring/documentation slice after scoped tests.
4. Commit only review fixes after rerunning scope tests.

## Verification

Run affected engine, connsdk, coordination, and connector tests plus focused race tests; then gofmt, targeted vet/build, connectorgen validation/surface-sync, individual non-full-suite project gates, and the GSD workflow verifier. Full `go test ./...` and monolithic `make verify` remain CI-owned by repository policy.
