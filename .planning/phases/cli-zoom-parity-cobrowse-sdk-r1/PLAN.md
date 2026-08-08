# Plan — Zoom Cobrowse SDK documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); slice:
  [#3945](https://github.com/polymetrics-ai/cli/issues/3945).
- Scope owner: Zoom's provider-owned **Cobrowse SDK** category only: four bounded direct reads,
  Zoom-local tests/fixtures, Zoom's derived endpoint projection, generated CLI/website docs, and
  this phase evidence.
- Required skills in scope: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-documentation`, and `no-mistakes` (loaded for the parent run and recorded here).
- GSD resolution: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review`; then the corresponding generated prompts.
  This provider-category phase is not registered by the official runtime, so it follows the
  documented inline manual-GSD fallback: plan, red, execute, verify, and review evidence remain
  here, and no forbidden role worker is spawned.

## Source audit — completed before RED

The source of truth is Zoom's current provider artifact, not the inherited ledger.

| Item | Evidence |
| --- | --- |
| URL | `https://developers.zoom.us/docs/api/cobrowse-sdk.md` |
| Retrieval | `2026-08-08T10:10:47Z` |
| HTTP / bytes | `200` / `11,697` |
| Artifact | OpenAPI `3.1.1`, API `2`, server `https://api.zoom.us/v2` |
| Reconciled operations | GET `/cobrowsesdk/live_sessions`, GET `/cobrowsesdk/past_sessions`, GET `/cobrowsesdk/sessions/{sessionId}`, GET `/cobrowsesdk/sessions/{sessionId}/users` |
| Ledger delta | `0` — every `provider_module=cobrowse-sdk` row matches method, origin-relative path, title, source URL, and module provenance |

The live and past session descriptions explicitly permit optional monthly `from` and `to` query
parameters. Both must be typed `date` query flags. The artifact describes `page_size` and
`next_page_token` only in response content, so neither becomes a CLI flag. The two session paths
require their declared `{sessionId}` path identifier. No other request parameters are declared.

## Locked implementation decisions

- Implement all four operations. There are zero `unsafe_or_disallowed` rows and no exclusions.
- Add four `rest_read` operations, all bounded to 1 MiB with `json_redacted` plus a Cobrowse-local
  sensitive field policy. It must redact the join-capable session pin, session/user identifiers,
  user display names, connection IDs, and IP addresses before CLI output.
- Expose `--from`/`--to` only on the two report reads and `--session-id` only on the two detail
  paths. Do not hand-author `page`, `per_page`, `limit`, `page_size`, or `next_page_token` flags.
- `surface-sync` owns derivable command metadata and the endpoint ledger; the reusable
  `surface-reconcile --notes-contains provider_module=cobrowse-sdk` foundation owns all four
  direct-read coverage rows.
- Regenerate CLI docs once and retain only Zoom/catalog deltas. The website has no hand-authored
  Zoom page, but its generated connector bundle/catalog is a required parity surface; regenerate
  it with `npm --prefix website run gen:catalog` and verify only `zoom` changed.

## TDD execution slices

1. **Plan checkpoint** — this evidence and source audit, before any Zoom production declaration.
2. **RED checkpoint** — change only the Zoom command-surface test and synthetic Cobrowse fixtures:
   exact aggregate expectations `18 → 22` covered, local blocked `1824 → 1820`, direct reads
   `13 → 17`, writes unchanged at `2`; execute against the pre-implementation bundle and preserve
   the literal failure. Commit and push this red state before production JSON/docs.
3. **GREEN checkpoint** — declare all four direct reads, command paths/flags, fixtures, source
   coverage, metadata/docs, and generated projections. Test exact routes/query contracts,
   redaction, no response-only pagination inputs, and required session IDs under an isolated fixture
   server. Run the scoped reconciler, validator, docs, website, and golden parity checks.
4. **Verify/review checkpoint** — build `pm`, run every route's help and safe live-read reachability
   with an environment-only synthetic token (expect Zoom `401`, not unknown command), record inline
   verification/review, commit/push, then update #3945 and #3915 with a clean next-category handoff.

## Verification plan

- RED and GREEN: `go test -count=1 ./internal/connectors/defs/zoom/...`.
- Surface: `go run ./cmd/connectorgen surface-sync --check`, `go run ./cmd/connectorgen validate
  internal/connectors/defs/zoom`, and full `go run ./cmd/connectorgen validate`.
- Runtime: targeted Zoom conformance, `go test -timeout 20m ./internal/connectors/conformance/...`,
  `go test -timeout 20m ./internal/connectors/commandrunner/...`, `go vet ./...`, and
  `go build -o <temporary>/pm ./cmd/pm`.
- Binary: `pm help zoom`, `pm zoom`, `pm zoom cobrowse-sdk`, all four route `--help` calls, then
  four safe GETs with a synthetic token expecting provider `401`, never a write.
- Docs: `pm docs validate --connectors-dir docs/connectors`; `npm --prefix website run gen:catalog`
  and `npm --prefix website run typecheck`; scoped docs/catalog/golden diff review.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.planning/phases/cli-zoom-parity-quality-management-r1/`
- `https://developers.zoom.us/docs/api/cobrowse-sdk.md`
