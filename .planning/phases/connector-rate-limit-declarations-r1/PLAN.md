# Provider-cited rate-limit declarations — R1 plan

## Objective

Ship the first bounded batch of provider-cited `rate_limits.json` declarations so the requester
admission machinery from #3877 begins pacing real outbound connector traffic. Include the first
production `defs.FS` embed pattern in the same slice.

## Scope

- Select roughly 20–30 connectors only from the authoritative completed provider-artifact sweep
  ledger at `/Users/karthiksivadas/karthik-agent-workspace/data/cli-provider-artifact-sweep-r1/ledger.json`.
  A declaration candidate must have `status: done` and its recorded live, enumerable provider
  artifact; `unknown` and `skipped` sweep records are never rate-limit candidates.
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

## Population and lifecycle gate

- The sweep ledger, not the 550 connector definition directories, is the authoritative population.
  On 2026-08-06 it contained 359 records: 281 `done`, 71 `unknown`, and 7 `skipped`; this is the
  live-file result and supersedes the earlier task snapshot that differed by one `unknown` record.
- A dead or retired provider surface is a lifecycle/deprecation finding, not a rate-limit
  `unknown`. Do not create a `rate_limits.json` or a rate-limit-ledger record for it. Capture it
  separately in `POPULATION-AUDIT.md`.
- A recheck found that `vercel` is absent from the current sweep ledger, rather than a `done`
  record. Its declaration and rate-limit ledger entry were removed; it is neither a rate-limit
  `unknown` nor a deprecation candidate. The retained 24 first-batch declarations all have a
  `done` record, a provider artifact, and `scope_in_current_defs: true`.

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

## Batch 2 — live, enumerable provider slice

The second bounded slice contains only the following `status: done` sweep records, each of which
has a live provider artifact and `scope_in_current_defs: true`: `7shifts`, `activecampaign`,
`aha`, `aircall`, `algolia`, `apify-dataset`, `appsflyer`, `ashby`, `assemblyai`, `bamboo-hr`,
`bitbucket`, `box`, `braze`, `brevo`, `cisco-meraki`, `commercetools`, `discord`, `google-ads`,
`mailchimp`, `mailgun`, `pagerduty`, `pipedrive`, `reddit`, `square`, and `twilio`.

The batch begins at `aha`, the stored resume point from batch 1, and selects the remaining
well-documented providers deliberately rather than treating all definition directories as a
population. Every declaration will be independently rejoined to the sweep ledger immediately
before it is written. A provider policy that depends on a secret token, an undeclared subject,
plan/tier not represented by the bundle, an unsupported subject type, endpoint dimensions not
represented by the bundle, or an unpublished replenishment shape is `unknown`; it is never
approximated.

The first-batch `defs.FS` wildcard already ships the optional file. This slice must not modify it,
nor any legacy `streams.json` throttle.
