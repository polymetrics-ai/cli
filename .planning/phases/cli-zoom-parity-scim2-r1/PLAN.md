# Plan — Zoom SCIM2 documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3942](https://github.com/polymetrics-ai/cli/issues/3942).
- Scope: Zoom's provider-defined **SCIM2** category only: all eleven documented operations,
  their exact typed CLI routes, generated Zoom docs/site output, fixture-backed lifecycle checks,
  any reusable foundations needed by their declared contracts, and this phase evidence.
- Required skills carried by the parent delivery: `golang-how-to`, `golang-cli`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review` on 2026-08-08. This provider-category phase
  is not registered by the official runtime, and the parent contract forbids role spawning; this
  is the documented inline manual-GSD fallback with explicit discussion/plan/RED/GREEN/verify/
  review evidence.

## Inline discuss-phase record

The provider's own SCIM2 category is a coherent slice: it contains Group and User SCIM resources
under one provider-owned category and one published artifact. It is independent of the already
landed Chatbot slice, except for the shared declarative executor. The category is selected before
any implementation because it has 11 bounded operations and no undocumented resource grouping is
introduced.

## Live artifact audit — completed before RED

The source is Zoom's current provider artifact, not the inherited endpoint ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/scim2.md` |
| Retrieval | `2026-08-08T13:33:09Z` |
| HTTP / bytes | `200` / `171,559` |
| SHA-256 | `ba86462a888677ea38a8bcc0557e9c4cf5809cd78fc6bc7655f85f79e5b27264` |
| Artifact | OpenAPI `3.1.1`, API server `https://api.zoom.us` |
| Ledger delta | `0` — exactly eleven `provider_module=scim2` rows already match method, path, title, and source URL |

The live `## Operations` section contains exactly these eleven actions:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/scim2/Groups` | List groups |
| POST | `/scim2/Groups` | Create a group |
| GET | `/scim2/Groups/{groupId}` | Get a group |
| DELETE | `/scim2/Groups/{groupId}` | Delete a group |
| PATCH | `/scim2/Groups/{groupId}` | Update a group |
| GET | `/scim2/Users` | List users |
| POST | `/scim2/Users` | Create a user |
| GET | `/scim2/Users/{userId}` | Get a user |
| PUT | `/scim2/Users/{userId}` | Update a user |
| DELETE | `/scim2/Users/{userId}` | Delete a user |
| PATCH | `/scim2/Users/{userId}` | Deactivate a user |

All actions document the `scim2` / `scim2:admin` scopes. The source declares `204 No Content` for
both deletes and group update, which must be asserted as status-only successes. Group/user SCIM
resources and SCIM PatchOp payloads contain provider-defined extensible objects; user creation and
update expose a large set of Zoom extension fields, including a provider-documented custom attribute.
The live artifact does not declare a standalone paging flag for the read commands, so none will be
hand-authored.

## Locked decisions

- Implement all eleven operations: four bounded `rest_read` / `direct_read` commands and seven
  `rest_write` / `direct_write` commands. No row is `unsafe_or_disallowed`, no duplicate is
  excluded, and no generic HTTP, shell, or SQL escape hatch is introduced.
- Use resource paths that mirror the provider: `scim2 groups list|get|create|update|delete` and
  `scim2 users list|get|create|update|delete|deactivate`. All mutations retain plan → no-network
  preview → explicit single-use approval → execute; both DELETE actions also require destructive
  typed confirmation. Status-only responses use `output_policy: none`.
- SCIM's provider server is `https://api.zoom.us`, whereas Zoom's ordinary bundle base ends at
  `/v2`. Build a reusable operation-scoped **direct-read** origin/auth foundation so SCIM reads
  reach exactly the declared root path and never inherit unrelated ordinary headers. The foundation
  must require paired `rest.base_url` and `rest.auth`, preserve the fixed endpoint path, and use
  the same secret-isolation principle as the existing direct-write override. It will also unblock
  documented read operations hosted on a distinct provider origin/base path.
- Build a reusable **named root JSON-object body** foundation. It may map only a declared
  `json_object` flag to exact `maps_to: "body"` on an operation-declared object body schema. It
  remains a fixed-operation typed input with normal request-size/schema validation; it is not a
  generic JSON flag or raw-body transport. This covers Zoom SCIM's documented extensible Group,
  User, and PatchOp resource shapes without silently dropping provider-defined fields.
- Give every SCIM operation a paired origin/auth declaration using a dedicated `scim2_base_url`
  config default and the ordinary Zoom bearer secret. This preserves the correct root origin while
  preventing a command from accidentally inheriting an unrelated operation's transport.
- Redact SCIM user/group/resource request values in previews, errors, and JSON results where
  provider fields carry personal or account data. Synthetic fixtures only are used; no credential,
  token, or token-derived value is printed or recorded.
- Run `connectorgen surface-sync`; never hand-author metadata it derives. Generate docs/site data
  repository-wide, then restore whole unrelated generated files and retain only Zoom-specific output.

## TDD execution

1. **Plan checkpoint** — commit this source audit, manual-GSD fallback, foundation decisions,
   target accounting, and verification plan before test or production changes.
2. **RED checkpoint** — add tests only and run them before any production change. They must prove:
   `27 → 38` executable rows, `1815 → 1804` locally blocked rows, `17 → 21` direct reads, and
   `5 → 12` direct writes; all eleven command paths are unknown before declaration; named root
   object mapping is unsupported; and a rest-read operation with declared paired root origin/auth
   still incorrectly reaches the ordinary base.
3. **GREEN foundations** — separately implement and test operation-scoped direct-read origin/auth
   plus the narrow named-root-object body mapping in engine, commandrunner, and static validation.
4. **GREEN connector** — declare the eleven operations, origin config, exact command groups/routes,
   body schemas, redaction, fixtures, endpoint coverage, generated docs, and website catalog.
   Reconcile only SCIM2 rows to `covered_by.direct_read` or `covered_by.direct_write`.
5. **Verify/review** — build `pm`; run base/group/every-command help and every mutation through
   isolated plan/preview/approval fixtures. Assert exact method/path/body/auth/status, deletion
   confirmation, no-body semantics, no invented paging input, and no leaked synthetic sensitive data.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable operations | 27 | 38 |
| Zoom-local implementable rows | 1,815 | 1,804 |
| Direct reads | 17 | 21 |
| Direct writes | 5 | 12 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- RED/GREEN focused tests for engine direct-read origin/auth, commandrunner root-object shaping,
  connectorgen static validation, Zoom bundle command preflight, app lifecycle, conformance/certify,
  and generated docs.
- `go run ./cmd/connectorgen surface-sync --check`, full validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=scim2`.
- Fresh binary help: `pm help zoom`, bare `pm zoom`, bare `pm zoom scim2`, and every exact SCIM2
  command `--help`.
- Fresh binary plan/preview/approval/execute for every mutation against isolated loopback fixtures;
  no real provider mutation or credential is used.
- Scoped CI-equivalent gates from `AGENTS.md`: vet, lint, docs, website typecheck, CLI transcripts,
  contract/surface/boundary/release checks. The full suite remains CI-owned.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/scim2.md`
