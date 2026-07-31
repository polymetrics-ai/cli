# Issue 3046 Gmail parity wave03 plan

## Scope

Parent issue: #3046. Capability subissues: #3047-#3053. Worker branch: `fm/cli-gmail-parity-wave03-r1`.

This wave is fixture-only. No live Gmail calls, credentials, provider writes, push, PR, /no-mistakes, VPS, Thaalam, or shared runtime behavior changes.

Implementation is Gmail-definition local except for generated docs/catalog/website/golden surfaces and one connectorgen validation-tooling fix that makes the documented focused gate `go run ./cmd/connectorgen validate internal/connectors/defs/gmail` validate a single bundle directory instead of misclassifying `fixtures/` and `schemas/` as connectors.

## GSD path and loaded skills

- `scripts/gsd doctor`: passed in this worktree.
- `scripts/gsd list`: passed; registry has no `programming-loop` command even though AGENTS.md references it.
- Attempted `scripts/gsd prompt programming-loop init --phase issue-3046-gmail-parity --dry-run`: failed with `unknown GSD command: programming-loop`.
- Fallback GSD evidence: generated `.planning/issues/gmail-parity-wave03/gsd-issue-rebootstrap-prompt.txt` with `scripts/gsd prompt issue-122-rebootstrap`; proceeding with manual GSD programming loop per `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`, `golang-spf13-viper`.
- CLI parity reference loaded: `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Source audit

Official sources re-audited:

- Gmail discovery: https://www.googleapis.com/discovery/v1/apis/gmail/v1/rest, revision `20260727`, version `v1`.
- Gmail REST reference: https://developers.google.com/workspace/gmail/api/reference/rest.
- Local artifact: `.planning/issues/gmail-parity-wave03/official-gmail-discovery-operations.json` (79 operations).

Pre-change local disposition from `internal/connectors/defs/gmail/api_surface.json`:

- total 79, covered streams 10, covered writes 35, excluded 34, operation rows 0.
- CSE operations were the 11 true excluded/N/A rows.
- The other 23 rows were legacy exclusions and needed executable coverage or blocked operation rows with exact evidence.

Post-change disposition:

- total=79, executable=61: 10 stream-covered operations, 40 typed write-covered operations, 11 operation-backed direct reads.
- blocked/planned=18: 5 email-path direct reads blocked on the shared direct-read identifier-safe path-variable validator, 2 CDC watch/stop controls blocked on shared CDC foundations (#2986/#2988), 11 CSE enterprise/admin rows.
- excluded/not-applicable semantic disposition=11 CSE rows; certified=0 live providers.

## Implementation slices

1. **Ledger and audit evidence**
   - Set `operation_ledger_version: 1`.
   - Partition all 79 official operations exactly once.
   - Keep CSE as not-applicable/admin operation rows (11 rows) with source-backed reasons.

2. **Typed writes**
   - Added `batch_modify_messages` and `batch_delete_messages` as closed-schema Gmail writes.
   - Added S/MIME certificate-management writes expressible by the existing JSON write dialect: `insert_smime_info`, `set_default_smime_info`, `delete_smime_info`.
   - Added sanitized fixtures for new write actions.

3. **Bounded direct and binary-like reads**
   - Added definition-owned `cli_surface.json` commands for single-resource/detail settings reads and attachment JSON payload reads.
   - Implemented 11 fixed official endpoints with `json_redacted` and operation-level byte caps/sensitivity metadata.
   - Marked 5 email-address path direct reads planned/blocked because the shared operation direct-read executor currently validates path variables with identifier-safe characters only; ETL and typed write templates still support those email path values.

4. **CDC/control-plane truthfulness**
   - Kept `watch` and `stop` blocked with operation rows until the shared CDC/changefeed foundations named in #3048 (#2986/#2988) define state/lab behavior.

5. **Docs/generated surfaces**
   - Updated Gmail `docs.md`, generated Gmail connector MANUAL/SKILL, all-connectors catalogs, website connector data, and CLI/help golden transcripts.
   - Verified `pm gmail`, `pm gmail --help`, and representative command help render safe metadata.

6. **Issue addendum**
   - Append an idempotent captain-policy addendum to #3046-#3053 using `gh-axi`, with actual post-change counts and fixture-only/uncertified status.

## Safety constraints

- No secrets in fixtures/docs/issues.
- No generic raw API commands, arbitrary paths, raw method/body, SQL write, shell, or file escape hatches.
- Reverse ETL remains plan -> preview -> explicit approval -> execute.
- Destructive writes have `confirm: "destructive"` and idempotency where the provider/contract supports it.
- No shared engine/runtime behavior changes.

## Expected local gates

1. `go run ./cmd/connectorgen validate internal/connectors/defs/gmail`
2. `go test ./internal/connectors/conformance -run 'TestConformance/gmail' -count=1`
3. `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
4. `go build ./cmd/pm`
5. `make connector-boundary`
6. `make verify`
7. `git diff --check`

## Orchestration decision

Cycle `plan`: `local_critical_path`. Firstmate already owns parent orchestration and this worker has one isolated worktree. The remaining Gmail work touches one connector's defs/docs/generated surfaces plus focused validation tooling, so spawning additional mutating workers inside this worktree would violate the isolation rule.
