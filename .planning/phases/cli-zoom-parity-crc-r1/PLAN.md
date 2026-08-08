# Plan — Zoom CRC documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); provider-owned slice:
  [#3937](https://github.com/polymetrics-ai/cli/issues/3937).
- Scope: every operation in Zoom's published **Conference Room Connector (CRC)** category,
  including declared direct reads, approval-gated direct writes, generated Zoom-only docs/site
  output, endpoint reconciliation, and GSD/TDD evidence.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- GSD provenance: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review` on 2026-08-08. This provider-category
  phase is not registered by the official runtime and the parent delivery contract forbids role
  spawning, so the work uses the documented inline manual-GSD fallback with explicit discussion,
  plan, RED, GREEN, verification, and review evidence.

## Inline discuss-phase record

CRC is Zoom's own published category. Its account settings, API connectors, managed rooms,
participant identifier, and room templates are provider-defined resources and remain one slice;
they are not a locally invented grouping. All twenty documented operations are in scope. No row
is excluded, and no row may use `unsafe_or_disallowed`.

The category has nine bounded sensitive reads and eleven approval-gated mutations. Seven mutation
responses are documented `204 No Content`; they remain real actions with status-only success, not
fake body responses. The private-key GET and regeneration PATCH are implemented with
`json_redacted` output and explicit private-key redaction. The three DELETE actions retain typed
destructive confirmation.

## Live artifact audit — completed before RED

The source was fetched afresh rather than trusting the inherited ledger.

| Item | Evidence |
| --- | --- |
| API URL | `https://developers.zoom.us/docs/api/crc.md` |
| Retrieval | `2026-08-08T20:10:59Z` |
| HTTP / bytes | `200` / `115,915` |
| SHA-256 | `a631ec0cc101a33df9b6483f772e26b334adc7ab8f6d265cbc6f48c863a8e2ba` |
| Artifact | OpenAPI `3.1.1`, API version `2`, server `https://api.zoom.us/v2` |
| Ledger comparison | exactly 20 local `provider_module=crc` rows; method, path, title, and source URL match (delta `0`) |

The source declares exactly these operations:

| Method | Path | Provider title |
| --- | --- | --- |
| GET | `/crc/managed_rooms/account_setting` | Get Cisco/Polycom Room Account Setting |
| PATCH | `/crc/managed_rooms/account_setting` | Update Cisco/Polycom Room Account Setting |
| GET | `/crc/api_connectors` | List API Connectors |
| POST | `/crc/api_connectors` | Create an API Connector |
| GET | `/crc/api_connectors/{connectorId}` | Get an API Connector |
| DELETE | `/crc/api_connectors/{connectorId}` | Delete an API Connector |
| PATCH | `/crc/api_connectors/{connectorId}` | Update an API Connector |
| GET | `/crc/api_connectors/{connectorId}/private_key` | Get an API Connector's private key |
| PATCH | `/crc/api_connectors/{connectorId}/private_key` | Update an API Connector's private key |
| GET | `/crc/managed_rooms` | List Managed Rooms |
| POST | `/crc/managed_rooms` | Create a Managed Room |
| GET | `/crc/managed_rooms/{deviceId}` | Get a Managed Room |
| DELETE | `/crc/managed_rooms/{deviceId}` | Delete a managed room |
| PATCH | `/crc/managed_rooms/{deviceId}` | Update a Managed Room |
| GET | `/crc/participant_identifier_code` | Get participant identifier code |
| GET | `/crc/room_templates` | List Room Templates |
| POST | `/crc/room_templates` | Create a Room Template |
| GET | `/crc/room_templates/{templateId}` | Get a Room Template |
| DELETE | `/crc/room_templates/{templateId}` | Delete a room template |
| PATCH | `/crc/room_templates/{templateId}` | Update a Room Template |

The artifact shows response-only pagination values for list responses. No `page`, `per_page`,
`limit`, cursor, or page-size flag is hand-authored; pagination remains declaration-derived.

## Locked implementation decisions

1. Declare nine bounded `rest_read` operations and eleven declared `rest_write` operations.
   All mutations stay behind plan → no-network preview → explicit single-use approval → execute.
2. Use exact, provider-closed request objects. The `account-setting` multipart form and the
   `managed-room` / `room-template` JSON objects accept only documented members; no generic JSON,
   HTTP, SQL, or shell write path is exposed.
3. Keep source operation paths in the ledger's declared form and let `surface-sync` / scoped
   reconciliation derive executable `/v2` endpoint metadata. Generated aggregate outputs are
   produced normally, then mechanically retained only where their semantic entry is Zoom.
4. Redact room identifiers, connector identifiers, room/template names, addresses, credentials,
   and `private_key` fields in previews and output. Tests use synthetic fixtures and never print
   a credential, token-derived value, or private key.
5. Treat `PATCH /crc/managed_rooms/account_setting`, `DELETE /crc/api_connectors/{connectorId}`,
   `PATCH /crc/api_connectors/{connectorId}`, `DELETE /crc/managed_rooms/{deviceId}`,
   `PATCH /crc/managed_rooms/{deviceId}`, `DELETE /crc/room_templates/{templateId}`, and
   `PATCH /crc/room_templates/{templateId}` as status-only actions. The three DELETE operations
   additionally require destructive typed confirmation.

## Required reusable foundation — camelCase path-variable derivation

The generated CLI standard is kebab-case (`--connector-id`), while Zoom's own endpoint templates
use camelCase (`{connectorId}`). `surface-sync` previously considered only a kebab-to-snake case
match, leaving a required path binding absent and making a command fail runtime preflight. This is
derived metadata, not a connector-author choice, so the foundation extends `surface-sync` to derive
the exact lower-camel path variable only when it is explicitly present in the declared operation
path. It does not infer query/body bindings, alter authored non-path mappings, or accept generic
paths. The foundation is developed and committed separately from CRC declarations; it benefits any
provider artifact using conventional camelCase path templates.

## TDD execution

1. **Plan checkpoint** — record the live artifact, source/ledger comparison, operation inventory,
   GSD fallback, and target accounting before bundle changes.
2. **RED checkpoint** — add only CRC command-surface/lifecycle tests and target-count bumps;
   run them against the current bundle and commit the failure verbatim before any production
   bundle edit.
3. **GREEN foundation** — deliver camelCase path-variable derivation as its own red/green commit,
   then rerun `surface-sync` to fill CRC's required path bindings.
4. **GREEN connector** — author `operations.json`, `cli_surface.json`, and ledger coverage;
   use `surface-sync` and scoped reconciliation instead of hand-editing derived metadata.
5. **Verify/review** — execute fixture lifecycle checks, build a fresh binary and run every CRC
   help route, run scoped validation/gates, record manual `verify-work` and `code-review`, then
   commit and push the coherent category slice.

## Target accounting

| Measure | Before | After |
| --- | ---: | ---: |
| Zoom executable endpoints | 123 | 143 |
| Zoom-local implementable rows | 1,719 | 1,699 |
| Direct reads | 61 | 70 |
| Direct writes | 58 | 69 |
| Binary downloads | 1 | 1 |
| Reverse-ETL writes | 2 | 2 |
| `unsafe_or_disallowed` Zoom rows | 0 | 0 |

## Verification plan

- Real `commandrunner.Preflight` coverage for all twenty exact `crc …` command paths.
- Fixture lifecycle coverage for seven 204 status-only actions, three destructive confirmations,
  and redacted private-key responses.
- `go run ./cmd/connectorgen surface-sync --check`, full connector validation, and scoped
  `surface-reconcile --check --notes-contains provider_module=crc`.
- Fresh binary `pm help zoom`, bare `pm zoom`, bare `pm zoom crc`, and every exact CRC route's
  `--help` output.
- Scoped test, vet, lint, docs, website, contract, surface, boundary, and release checks from
  `AGENTS.md`; CI owns the full repository suite.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `https://developers.zoom.us/docs/api/crc.md`
