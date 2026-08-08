# Plan — Zoom Clips documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3936](https://github.com/polymetrics-ai/cli/issues/3936).
- Scope: Zoom's published **Clips** artifact only: all twenty-one documented endpoints, their
  typed command contracts and fixtures, necessary reusable engine foundations, generated Zoom
  docs/site output, endpoint reconciliation, and GSD/TDD evidence.
- Required skills carried by the parent delivery: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review` on 2026-08-08. This provider-category phase
  is not registered by the official runtime and the parent contract forbids role spawning, so
  this is the documented inline manual-GSD fallback with explicit discussion, plan, RED, GREEN,
  verification, and review evidence.

## Inline discuss-phase record

Clips is Zoom's own published category. Its operations naturally group as clip listing/detail,
collaboration, comments, chapters, ownership transfers, download, and the provider's clip-file
upload sequence; splitting those resources into a locally invented category would obscure the
provider surface. All twenty-one endpoints are included. There are no duplicate exclusions and
no `unsafe_or_disallowed` rows.

The previous worker's five uncommitted files in worktree `.../26/cli` were inspected read-only
before this phase. They are an already-landed Healthcare candidate: the two operations and three
CLI paths are present on this branch through `1d260747c`, and the formerly untracked
`writes.json` action is tracked here. The old diff was therefore stale work, not a Clips input;
nothing was blindly copied or deleted. That durable work remains carried by this branch.

## Live artifact audit — completed before RED

The source was fetched afresh rather than trusting the inherited ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/clips.md` |
| Retrieval | `2026-08-08T17:23:43Z` |
| HTTP / bytes | `200` / `57,603` |
| SHA-256 | `ea22469a6432b79f2bc09ad6345419d737577e53ca170a70e7855327c011d764` |
| Artifact | OpenAPI `3.1.1`, API version `2`, server `https://api.zoom.us/` |
| Ledger comparison | exactly 21 local `provider_module=clips` rows; method, path, title, and source URL match (delta `0`) |

The artifact contains exactly these twenty-one provider operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/clips` | List all clips |
| POST | `/clips/starred` | Batch star clips |
| GET | `/clips/{clipId}/collaborators` | Get collaborators of a clip |
| POST | `/clips/{clipId}/collaborators` | Share a clip with new users |
| DELETE | `/clips/{clipId}/collaborators` | Remove the collaborator from a clip |
| GET | `/clips/{clipId}/comments` | List clip comments |
| DELETE | `/clips/{clipId}/comments/{commentId}` | Delete a comment |
| GET | `/clips/{clipId}/download` | Download a clip |
| GET | `/clips/{clipId}` | Get a clip |
| DELETE | `/clips/{clipId}` | Delete a clip (soft delete) |
| PATCH | `/clips/{clipId}` | Update a clip's basic info |
| GET | `/clips/{clipId}/chapters` | Get Clip Chapters |
| POST | `/clips/{clipId}/chapters` | Create Clip Chapters |
| POST | `/clips/{clipId}/duplicate` | Duplicate Clips |
| PATCH | `/clips/{clipId}/share_settings` | Update clips share setting |
| POST | `/clips/transfers` | Transfer clips owner |
| GET | `/clips/transfers/{taskId}` | Transfer task status check |
| POST | `/clips/files` | Upload clip file |
| POST | `/clips/files/multipart` | Upload clip multipart files |
| POST | `/clips/files/multipart/upload_events` | Initiate and complete the multipart file upload for a clip |
| POST | `/clips/files/tmp` | Temporary file upload API for Clips |

The source has response pagination fields but no request query-parameter sections for the three
list endpoints. No `page`, `per_page`, `limit`, page-size, or cursor flag will be hand-authored.
The API version is `2`; executable Zoom paths use the connector's fixed `/v2` API root while
the source ledger retains source-normalized paths until generated reconciliation updates it.

## Required foundations, kept closed and declaration-owned

1. **Root JSON array direct writes.** `POST /clips/{clipId}/collaborators` accepts an array,
   not an object. Extend the typed direct-write contract and CLI flag dialect with one exact
   `json_array` root-body option. It must accept exactly one JSON array, validate it against the
   operation's closed root schema, forbid mixing it with dotted body fields, and retain the
   ordinary plan → preview → single-use approval → execute gate. It is not a generic raw-body
   escape hatch.
2. **Provider-bound bearer redirect for binary download.** `GET /clips/{clipId}/download`
   explicitly returns a 302 and instructs callers to retain the Authorization header across the
   provider redirect. The existing binary executor correctly strips all credentials on an open
   cross-host hop; add a separate, declaration-owned exception which may preserve only a bearer
   `Authorization` header for HTTPS hosts inside an exact declared provider suffix, with a finite
   hop cap. The Zoom declaration is limited to `zoom.us`; default and custom-auth header stripping
   remains unchanged for every other download.
3. **Operation-level bounded base64 image upload.** `POST /clips/files/tmp` specifies a JSON
   binary string, accepts only PNG/JPEG/GIF/JPG images below 2 MB, and requires the same
   bearer-preserving Zoom redirect as file uploads. Reuse the existing approved local-file,
   base64 canonicalization policy through a closed `rest.base64_upload` declaration: the local
   source path never reaches the wire, decoded/encoded limits apply, source-name and image-media
   constraints are declaration-owned, and execution is bound to the previewed snapshot.

The two clip-video uploads use the existing closed multipart declaration and the existing
bearer-preserving mutation redirect foundation. Their declared `.mp4`/`.webm` extension and
upper bounds are enforced. The source's minimum 5 MB *except final part* cannot be expressed as
a static file constraint because this endpoint exposes no final-part indicator; rejecting a small
final part would contradict the provider contract. The declared `upload_context` is sent exactly
as documented so Zoom can decide that stateful condition.

`POST /clips/files/multipart/upload_events` has two documented `oneOf` request arms. It will be
represented as separate named `initiate` and `complete` commands sharing the one source endpoint,
so each direct-write contract has a closed concrete body and surface reconciliation records the
two-command coverage. Likewise partial and full transfers are separate named commands where that
preserves their mutually conditional documented payloads.

## Locked decisions

- Implement six bounded redacted JSON direct reads, one bounded binary download, and sixteen
  approval-gated direct-write commands covering fourteen write endpoints. The two extra commands
  are the source's multipart-event and transfer discriminated variants, not duplicate endpoints.
- All documented `204 No Content` actions are destructive/status-only actions, gated by typed
  destructive confirmation where they delete a collaborator, comment, or clip.
- JSON bodies are closed schemas. Source text constraints use the existing schema subset where
  representable (array cardinality, enum, numeric bounds, date/time format, and bounded patterns)
  without inventing a generic client input format.
- Read and mutation response fields that can contain clip IDs, owner/user IDs, names, email,
  comments, passcodes, URLs, upload contexts, task IDs, file IDs, source paths, and tokens are
  redacted. Fixtures are synthetic and no credential-derived value is printed.
- Derived metadata comes only from `surface-sync` and reconciliation. Docs/site are generated
  repository-wide, unrelated output is restored, and only Zoom output is retained.

## TDD execution

1. **Plan checkpoint** — commit this live-source audit, foundation decisions, manual GSD fallback,
   prior-worker disposition, and target accounting before test or production changes.
2. **RED checkpoint** — add only Clips command-surface tests plus foundation tests. Capture the
   current `102 → 123` executable, `1,740 → 1,719` Zoom-local, `55 → 61` direct-read, and
   `42 → 58` direct-write target failures; prove all 21 native paths are unknown via the real
   command runner; prove array-root, authorized binary redirect, and base64 operation upload are
   currently unavailable.
3. **GREEN foundations** — deliver each foundation in its own red/green commit sequence, preserving
   ordinary redirect stripping and object-body semantics in regression tests.
4. **GREEN connector** — declare every Clips operation and command, exact fixtures, source bounds,
   redaction, and approval semantics; reconcile only the Clips rows; regenerate derived metadata,
   docs, and website output.
5. **Verify/review** — run real fixture lifecycle tests, fresh binary base/group/every-command help,
   source/ledger checks, generated-output locality checks, scoped local gates, inline verify-work,
   and manual code review.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable endpoints | 102 | 123 |
| Zoom-local implementable rows | 1,740 | 1,719 |
| Direct reads | 55 | 61 |
| Direct writes | 42 | 58 |
| Binary downloads | 0 | 1 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN real-commandrunner preflight and fixture lifecycle tests for all 21 source endpoints
  and all 23 concrete command contracts.
- Foundation regression tests for root-array validation, bearer retention only inside declared
  suffixes, cross-origin credential stripping otherwise, and bounded image base64 source snapshots.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=clips`.
- Fresh binary `pm help zoom`, bare `pm zoom`, bare `pm zoom clips`, and every exact command
  `--help` route.
- Scoped CI-equivalent vet/lint/docs/website/CLI/contract/surface/boundary/release gates from
  `AGENTS.md`; the full repository suite remains CI-owned.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/clips.md`
