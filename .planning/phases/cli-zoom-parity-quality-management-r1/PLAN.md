# Plan — Zoom Quality Management documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); slice:
  [#3943](https://github.com/polymetrics-ai/cli/issues/3943).
- Scope owner: Zoom's provider-owned **Quality Management** category only: five bounded direct
  reads, one typed interaction-creation action, Zoom-local tests/fixtures, Zoom's derived endpoint
  projection, generated Zoom documentation, and this phase evidence.
- Required skills in scope: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-documentation`, and `no-mistakes` (loaded for the parent
  run and recorded here for this slice).
- GSD resolution: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review`; then the corresponding generated prompts.
  The provider-category phase is not registered by the official runtime, so this uses the same
  inline manual-GSD fallback as the completed Healthcare slice: plan, red, execute, verify, and
  review evidence remain here; no forbidden worker role is spawned.

## Source audit — completed before RED

The source of truth is the provider's own artifact, not the inherited ledger.

| Item | Evidence |
| --- | --- |
| URL | `https://developers.zoom.us/docs/api/quality-management.md` |
| Retrieval | `2026-08-08T09:12:11Z` |
| HTTP / bytes | `200` / `40,987` |
| Artifact | OpenAPI `3.1.1`, API `2`, server `https://api.zoom.us/v2` |
| Reconciled operations | GET `/qm/automated_evaluations`, GET `/qm/evaluation`, GET `/qm/evaluation/{evaluationId}`, GET `/qm/interactions`, POST `/qm/interactions`, GET `/qm/interactions/{interactionId}` |
| Ledger delta | `0` — all six `provider_module=quality-management` rows match method, origin-relative path, title, source URL, and module provenance |

The five GET operations declare no request-parameter section in the live artifact. The list response
examples mention pagination fields, but those fields are response data and are **not** hand-authored
CLI flags. The two detail routes expose only their documented path identifiers.

The POST body is fully enumerable: required `download_url`; optional `direction`, `disposition`,
`primary_language`, `queue_id`, and `start_time`; and optional `interaction_info` whose required
member, when the object is supplied, is `channel_type`. Its other documented scalar members are
`agent_email`, `agent_id`, `consumer_name`, `from`, and `to`. Explicit nested scalar flags map to
that closed body shape; no raw JSON flag or generic HTTP escape hatch is needed.

## Locked implementation decisions

- Implement all six operations. There are zero `unsafe_or_disallowed` rows and no exclusions.
- Add five `rest_read` operations, all bounded to 1 MiB with `json_redacted`. The output policy
  already redacts token/download-shaped fields; connector-local `sensitive_policy.redact_fields`
  additionally removes Quality Management identifiers and personal-contact fields from returned
  data before CLI output.
- Expose only `--evaluation-id` and `--interaction-id` for the detail reads. List reads take no
  provider-invented paging/date filters.
- Implement POST as a typed high-risk `create_quality_management_interaction` reverse-ETL action,
  with the exact closed record schema described above. It remains plan → preview → explicit approval
  → execute. The request's `download_url` and `interaction_info` are redacted in generic write
  errors; synthetic tests never use a real download URL or contact value.
- `surface-sync` owns derivable command metadata and the endpoint ledger. The reusable
  `surface-reconcile --notes-contains provider_module=quality-management` foundation owns only the
  five direct-read coverage rows; the POST row is linked by the typed write declaration.
- Generated documentation is regenerated once and scoped back to Zoom/catalog deltas only. No
  website-specific Zoom page exists, so `website/**` is not applicable.

## TDD execution slices

1. **Plan checkpoint** — this evidence and source audit, before any Zoom production declaration.
2. **RED checkpoint** — change only the Zoom command-surface test and synthetic fixture coverage:
   exact aggregate expectations `12 → 18` covered, local blocked `1830 → 1824`, direct reads
   `8 → 13`, writes `1 → 2`; execute it against the pre-implementation bundle and preserve the
   literal failure in `TDD-LEDGER.md`. Commit and push this red state before production JSON/docs.
3. **GREEN checkpoint** — declare all five direct reads, the closed typed POST action, command
   paths, fixtures, source coverage, metadata/docs, and generated projection. Test each GET path,
   redaction, absence of response-only flags, POST body, and `201` success under an isolated
   fixture server. Run the scoped reconciler, validator, and generated-doc parity checks.
4. **Verify/review checkpoint** — run the scoped Go checks, build `pm`, run every route’s help and
   safe live-read reachability with a synthetic environment token (expect Zoom 401, not unknown
   command), run the write route only in preview, record inline verification/review, commit/push,
   and update #3943 and #3915 with a clean next-category handoff.

## Verification plan

- RED and GREEN: `go test -count=1 ./internal/connectors/defs/zoom/...`.
- Surface: `go run ./cmd/connectorgen surface-sync --check`, `go run ./cmd/connectorgen validate
  internal/connectors/defs/zoom`, and full `go run ./cmd/connectorgen validate`.
- Runtime: `go test -timeout 20m ./internal/connectors/conformance/...`, `go test -timeout 20m
  ./internal/connectors/commandrunner/...`, `go test -count=1 -timeout 20m ./internal/cli/...`,
  `go vet ./...`, and `go build -o <temporary>/pm ./cmd/pm`.
- Binary: `pm help zoom`, `pm zoom`, `pm zoom quality-management`, every route `--help`, then
  all five safe GET routes with a synthetic token expecting provider `401`, never a live mutation.
  POST is proven by fixture `201` plus plan/preview only.
- Docs: `pm docs validate --connectors-dir docs/connectors`; scoped docs/catalog/golden diff review;
  bare namespace help exits successfully and missing required IDs remain usage errors.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.planning/phases/cli-zoom-parity-healthcare-r1/`
- `https://developers.zoom.us/docs/api/quality-management.md`
