# CLI PM Broker Client Contract R1

## Objective

Implement the independent CLI-side PM Broker OpenAPI HTTP/JSON `/v1` client/transport foundation on top of the integrated synthetic contract fixtures from PR #594.

## GSD and skills evidence

- GSD adapter: `scripts/gsd doctor` succeeded.
- Required GSD command path attempted: `scripts/gsd prompt programming-loop init --phase issue-585-pm-broker-client-contract --dry-run` returned `unknown GSD command: programming-loop`, so this slice uses the documented repo-local manual fallback.
- Manual fallback prompt captured with `scripts/gsd prompt gsd-quick "CLI PM Broker OpenAPI /v1 client contract lane"`.
- Issue read: `gh-axi issue view 585 -R polymetrics-ai/cli --full`.
- Dependency read: `gh-axi pr view 594 -R polymetrics-ai/cli --full` and `internal/pmbroker/contract/v1` fixtures/client.
- Required routing read: `.agents/agentic-delivery/references/required-skills-routing.md` and `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-design-patterns`, `golang-structs-interfaces`.

## Scope

- Extend `internal/pmbroker/contract/v1` with a typed HTTP/JSON client that consumes the existing fixture-backed `/v1` contract shapes.
- Support loopback HTTP and remote/container HTTP endpoints through the same request builder and custom round-tripper seam for deterministic tests.
- Add explicit internal authentication and correlation seams without exposing generic request, raw JSON/body, arbitrary URL/header, gRPC, socket, SQL, shell, or provider payload escape hatches.
- Enforce endpoint safety: no credentials/userinfo, query, fragment, unsupported schemes, or unsafe Host/Origin assumptions.
- Add transport coverage for compatibility negotiation, exact HTTP 426 `incompatible_contract_version`, bounded pagination, idempotency headers, immutable execution-plan digest headers, structured safe errors, rate-limit metadata, and redacted diagnostics.

## Non-goals and safety

- Do not touch `internal/pmbroker/domain.go` or profile/context validation behavior.
- Do not answer or bypass PMB-005/PMB-006 or any paused profile/context validation decisions.
- No live PM Broker/provider/GCP/VPS resources, real credentials, service-account JSON keys, customer data, raw-secret export, arbitrary authenticated HTTP, generic JSON/request escape hatches, SQL, shell, or runtime plugins.
- No public gRPC path or divergent Unix socket / named pipe semantics.
- No new dependencies.

## Implementation plan

1. Add failing synthetic HTTP client tests around the existing `contract/v1` package fixtures: endpoint validation/redaction, loopback/remote transport parity, auth/correlation seams, Host/Origin rejection, pagination/idempotency/digest, structured errors, rate limits, no gRPC/socket/generic escape.
2. Implement the smallest internal typed HTTP client surface and fake-broker handler enhancements needed to satisfy those tests while preserving the existing fake client API.
3. Run targeted package tests, `gofmt`, and broader deterministic Go checks as practical.
4. Commit the green slice on `fm/cli-pm-broker-client-contract-r1` for the sub-PR base `integration/pm-broker-production-program`.

## Review-fix plan

- GSD review-fix preflight on 2026-07-28: `scripts/gsd doctor` succeeded; `scripts/gsd prompt programming-loop init --phase pmbroker-contract-review-fixes --dry-run` returned `unknown GSD command: programming-loop`, so the documented manual-GSD fallback remains active.
- Manual fallback prompt: `scripts/gsd prompt gsd-quick "PM Broker contract review findings fix round"`.
- Required skills refreshed: `gsd-programming-loop`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`.
- Verified findings to fix: client compatibility negotiation must reject unsupported configured client versions during preflight; execution-plan responses must be bound to the submitted request; connector-connection response validation must accept contract enums rather than only the synthetic fixture; issue-first guard must not be bypassable by arbitrary `fm/*` branch names.
- Focused verification boundary: `gofmt` on touched Go files and `go test ./internal/pmbroker/contract/v1` only.
