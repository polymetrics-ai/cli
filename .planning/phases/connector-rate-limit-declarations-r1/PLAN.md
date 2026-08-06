# Provider-cited rate-limit declarations — R1 plan

## Objective

Ship the first bounded batch of provider-cited `rate_limits.json` declarations so the requester
admission machinery from #3877 begins pacing real outbound connector traffic. Include the first
production `defs.FS` embed pattern in the same slice.

## Scope

- Select roughly 20–30 connectors from the completed provider-artifact sweep.
- Research each provider's current rate-limit policy from a provider-controlled HTTPS source; the
  sweep ledger URL is only a documented starting point.
- Add one closed-schema `rate_limits.json` per selected connector. A policy must use an existing,
  non-secret `spec.json` property whose semantic subject matches the provider policy.
- Use explicit `unknown` records where a provider's public policy cannot be enumerated or mapped
  safely to the bundle's scope contract.
- Update `internal/connectors/defs/defs.go` with `*/rate_limits.json` once declarations exist.
- Maintain the ignored resumability ledger and `PROGRESS.md` after every completed connector.

## Non-goals

- No estimated, inferred, or undocumented quotas.
- No `streams.json` `base.rate_limit` additions or changes.
- No credentials, live connector checks, new dependencies, generic request surfaces, CLI help, or
  website/docs changes.
- This is a bounded foundation rollout, not a claim to cover every researched connector.

## Scope guard and ownership

This is the dedicated cross-connector rate-limit-declaration foundation slice requested by the
delivery brief. Its allowed production paths are `internal/connectors/defs/defs.go` and the named
`internal/connectors/defs/<connector>/rate_limits.json` files. The external, gitignored progress
ledger is required resumability state. No unrelated connector migration files may change.

## TDD / verification plan

1. Research a provider policy and verify its source URL, retrieval date, rate shape, and matching
   non-secret scope property before writing its declaration.
2. Write the first declaration **before** changing `defs.go`; run
   `go test ./internal/connectors/engine -run TestProductionDefinitionsEmbedEveryRateLimitDeclaration`
   as the expected red evidence. The test proves a source-tree declaration is not shipped until
   the optional wildcard is present.
3. Add the `*/rate_limits.json` embed pattern, then re-run the same test as green evidence.
4. Validate every declaration with `jq`, `go test ./internal/connectors/engine`,
   `go run ./cmd/connectorgen validate`, and
   `go run ./cmd/connectorgen surface-sync --check`.
5. Run the changed-package test and selected repository verification gates separately; do not run
   the monolithic suite under a per-command timeout. Record the exact outcomes in
   `VERIFICATION.md`.

## Required skills and GSD path

- Loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and
  `no-mistakes`.
- GSD commands resolved through `scripts/gsd`: `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`.
- Inline/manual fallback: this worker is operating outside Pi and the current worker contract
  forbids spawning lifecycle roles, so the generated prompts are executed inline with these phase
  artifacts as the durable record.

## Commit checkpoints

1. Plan and resumability ledger checkpoint.
2. First green declaration + production embed checkpoint.
3. Complete first batch + validation checkpoint.
4. Any review-fix checkpoint.
