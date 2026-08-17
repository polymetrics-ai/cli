# GitHub certification suite r1

## Task Delivery Header

- Issue: Refs #3993 — GitHub Certification: execute the complete implemented surface in one live run; Refs #4015 — Production MVP certification.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A verified direct PR is open against the stated base and its API-reported base exactly matches this header.
- Working branch: `fm/cli-github-certification-suite-r1`.
- Task: Generate GitHub certification candidates and exhaustive command accounting from its declared CLI surface; distinguish provider refusal from declaration/runtime product defects; preserve assertion-bearing live candidates as narrow named overlays; and make the complete sweep artifact deterministic and checkable. Do not emit accepted evidence until PR #4198 supplies `http_exchanges`.
- Verification: Focused red/green Go tests; regenerated sweep artifact and deterministic `--check`; a post-schema candidate assertion sabotage proving that operation's own case turns red then restores; built-binary live certification only for eligible declaration-owned candidates; consumer-package tests; and the repository's formatting, lint, docs, generated-file, boundary, and connector gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Candidates derive from declared surface rather than a hand-authored list | live | The generator loads `cli_surface.json`; removing an implemented surface command changes the generated accounting, while the legacy 23-entry assertion overlay can only contribute assertions to a generated command of the same path. |
| Every declared command has exactly one accountable status | live | The generated GitHub sweep contains 1,571 unique command paths, status totals sum to 1,571, and its validator rejects duplicate, omitted, or affirmatively passed unexecuted entries. |
| Provider refusals and product defects differ | live | Provider-refusal records require status/reason and remain non-pass; the generic required REST path-parameter comparison emits a product-defect finding for `releases assets view` and its non-required `--asset-id`. |
| Certified operation assertions prove a produced value | live | Only a generated candidate linked to a declaration-owned `/response/...` assertion is eligible for the existing live direct-read stage; its post-schema impossible assertion makes that case fail, then restoration passes. |
| Evidence promotion remains honest while #4198 is open | fake | Accepted evidence requires `http_exchanges`; PR #4198 is still open and no fake or partial proof record is created. The artifact records this as a concrete non-pass dependency. |

## Scope and ownership guard

- Target connector: GitHub only.
- Owned paths: `cmd/connectorgen/**` for generic surface-derived certification generation, its tests, the generated GitHub certification sweep artifact, GitHub's certification definition only when a narrow assertion overlay needs documentation, and this phase's evidence.
- Not owned: PR #4198 transport capture, PostgreSQL, connector engine/runtime execution changes, broker/MCP/UI work, credential scope changes, and any new generic write/HTTP/SQL tool.
- Boundary rule: generator code must operate on a loaded bundle and never name GitHub in shared Go. GitHub-specific fixture/credential facts stay in its definition or planning evidence.

## GSD and skills record

- Resolved lifecycle: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`; `plan-phase`; `execute-phase`; `verify-work`; and `code-review`; generated prompts with `--auto` / inline execution; and `go run ./cmd/agentcontractgen check` all pass before production edits.
- Inline/manual fallback: this autonomous task and the canonical single-worker delivery contract prohibit GSD-role spawning. The worker records discussion, plan, red/green ledger, verification, and review artifacts inline.
- Skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-documentation`.
- CLI/docs parity: `connectorgen` gains a developer generator subcommand, not the shipped `pm` command surface. Its own help and tests change; `pm` help/manual/website sources are explicitly unchanged and checked as such.

## TDD delivery slices

1. **RED — source-driven accounting contract.** Add generator tests that fail because the current code has no full-surface candidate artifact, no exact-one status accounting, and no `releases assets view` declaration/runtime mismatch finding.
2. **GREEN — deterministic generation and checking.** Add a generic `connectorgen certification-sweep` command that loads a bundle's `cli_surface.json`, derives one candidate/accounting record per command, adds only existing assertion overlays by matching declared command path, validates exhaustive status totals, and writes/checks a deterministic generated artifact.
3. **GREEN — defect and non-pass semantics.** Compare required REST path parameters with their generated CLI flags. Emit product defects separately from provider-refusal records; classification never presents unexecuted commands as pass and supplies a concrete reason for every status.
4. **FAILURE PROOF — operation assertion.** After schema compilation, scratch-corrupt one certified operation's produced-value assertion, run its own certification case red, restore the exact declaration, and rerun green. The sabotage is not committed.
5. **LIVE SWEEP — bounded and resumable.** Build the current binary and run only eligible declaration-owned candidates serially with the disposable identity. Record provider refusals separately and restore/remove all `pm-cert-` resources. If the dependency makes accepted evidence impossible, retain only sanitized non-pass operational evidence.
6. **VERIFY / REVIEW / DELIVERY.** Regenerate and check artifacts, run affected packages plus consumers, all local repository gates individually, review the diff inline, commit, push, open a direct PR, and API-read the base.

## Commit checkpoints

- Planning and fresh RED evidence.
- Green generator, generated GitHub sweep, and failure demonstration.
- Verification/review fixes and PR-ready evidence.
