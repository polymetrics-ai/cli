# Gong source audit — 2026-08-24 UTC

## Locked official source

- Source: `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`.
- Retrieved credential-free at `2026-08-23`.
- OpenAPI `3.0.1`; info version `V2`; 59 paths; 69 operations.
- Exact artifact: 453,797 bytes; SHA-256
  `294bf80b28e773d66a30bd0a8e76344140b17cf0225803e759d1e112b6b1fa13`.
- Canonical sorted `(method, path, operation_id, deprecated)` inventory fingerprint:
  `46cc30f319526572258d56293795c9036ea11b0ebee16919c7bb498b020f61e5`.
  The committed source lock produces the identical fingerprint.
- Method distribution: DELETE 3, GET 29, PATCH 1, POST 28, PUT 8.

## Exact eight-surface inventory

| Execution surface | Count | Declaration-owned mapping |
| --- | ---: | --- |
| ETL | 12 | Exact stream bindings and implemented `etl` commands. |
| Direct read | 30 | Implemented bounded CLI commands; every row has the exact API-surface endpoint. |
| Direct write | 27 | Every provider mutation has one implemented named command and exact API operation; typed confirmation remains required where declared. |
| Reverse ETL | 27 | The same 27 declaration-selected actions execute only through plan → preview → approval → execute. |
| Binary download | 0 | The official document declares no binary response operation. |
| Binary upload | 3 | `PUT /v2/calls/{id}/media`, `POST /v2/crm/entities`, and `POST /v2/targets/{targetId}/assignments`; named multipart actions with bounded files and approval-digest binding. |
| Flow | 1 application workflow | Gong is the capability-backed source in the harness's capture ETL → local query `flow_roundtrip`; no Gong-named flow-runtime branch or provider-flow mutation is invented. |
| Schedule | 1 application workflow | The same bounded Gong-backed flow is created, listed, installed into an isolated crontab file, fired, inspected, removed, and residue-checked by `schedule_roundtrip`; no provider scheduler API is assumed. |

The primary source-map class remains `direct_write` for the three multipart operations because
they are also fixed write actions; their binary-upload capability is recorded separately above.
No destructive provider operation is omitted: the three DELETE rows are implemented commands
with typed destructive confirmation.

## Output and foundation audit

- Gong direct reads declare no `sensitive_policy.redact_fields`. Their policy requires every
  ordinary provider value, including one that happens to equal configured credential material, to
  remain intact. Only explicitly declared secret output fields may be represented by an explicit
  marker. Shared collision masking currently violates this rule across provider-output projections;
  provider-neutral foundation issue #4321 owns its red/green correction. Connector docs and the
  focused surface test forbid wording or declarations that imply ordinary response-field redaction.
- The reconciled `origin/main` tree `6410fe59c` contains the final squashed structured-body,
  source-import, typed-header/binary/status/text, action-binding, and declaration-route heads.
  The branch merge preserves that shared tree and retains only Gong-owned declarations and
  evidence in the PR diff. Focused Gong multipart conformance must be re-run against this tree.
- PR #4335 is merged at `8127de418`. The Gong source lock now uses its v3 document form: the
  single `gong-v2` document retains all 69 existing rows, the fixed `?version=` URL is explicitly
  `identity_query: true`, and its queryless capture citation, digest, byte count, and provider
  version are all locked. The fresh scoped import traversed that declared query and parsed the
  official document; it therefore no longer fails URL policy.
- The scoped import is deliberately **not** marked green: its next failure is provider-neutral
  source-import preflight, at `GET /v2/all-permission-profiles` parameter 0: the official
  `workspaceId` required query input has schema `{ "type": "string" }` with no `maxLength`, so
  `sourceValidateLengthBounds` returns `unbounded request schema string has no maxLength`. The
  runtime already has a common bounded input boundary classification, but the earlier
  descriptor-building stage rejects the ordinary provider string before it can preserve and
  classify the contract. The required foundation is to retain such source-declared common-bound
  inputs through descriptor projection (or emit a typed source-bound gap), without inventing a
  Gong maximum, bypass, or provider branch. Until then no canonical Gong descriptor or derived
  required-input projection can be claimed.
- A fresh initialized project with no credential ran the built `pm` binary against all 69
  implemented Gong command paths: 30 direct reads, 27 reverse-ETL writes, and 12 ETL streams.
  Every command reached `missing --credential`; none returned `unknown command`, a partial-command
  block, or an API-surface preflight block. No request reached Gong.

## Live certification boundary

An approved disposable credential reference is available through a non-echoing local secret store.
It is used only at point of execution through stdin/environment delivery to the persisted App and
the repository certification harness; its values, account identifiers, and provider payloads are
never written to evidence. Gong agentic endpoints are excluded because they consume paid credits.

## Live bounded certification results

- The persisted App credential path passed its declaration-owned authentication check without
  emitting any credential value. The rebuilt binary SHA-256 was
  `62b3018f956abcd0537f7944154bbabc497c7b4344b187885504a34f5cdcf075`.
- A bounded `users list --limit 1` ETL invocation returned one record. The ordinary typed direct
  read `users extensive --max-bytes 1048576` returned HTTP-200-classified success with a page
  context. Its scoped external-proof run passed preflight, credential test, and direct read with
  six pass stages, one documented skip, and zero leaked resources. The proof remains only in the
  temporary certification project; its non-secret SHA-256 is recorded in `TDD-LEDGER.md`.
- The built CLI rejected a missing required `--call-id` before provider I/O and rejected an
  addressable `--page` for cursor navigation before provider I/O. These are the required-input and
  pagination boundary checks; no provider payload was retained.
- The repository `--full` harness made no writes and no agentic request. It observed 16 bounded
  ETL records, seven passing ETL append cells, five failing append cells (`calls`,
  `library_folders`, `flows`, `flow_folders`, and `permission_profiles`), 19 bounded rate-limit
  events, and zero leaks. A valid bounded `calls list` attempt with both typed date bounds still
  received HTTP-404 classification. These five stream cells therefore remain uncertified; the
  report does not turn them into partial commands.
- `GET /v2/targets` requires `workspaceId` in the current official document. The generated
  `targets list` flag is not yet marked required, and a live call without it returned only
  HTTP-400 classification. `params-import` reconciled the three multipart operation parameter
  blocks (17 scanned, then zero drift), but the required direct-read flag must be projected by
  `surface-sync`. That command is blocked because the canonical source descriptor cannot be
  generated while source-import rejects the ordinary unbounded input before projection. No manual
  flag was authored.
- There are no declaration-owned, self-cleaning Gong write pairings for safe live create,
  readback, and cleanup. All 27 provider writes stay implemented and certification-visible as
  unassessed; none was sent. The three multipart writes retain generic conformance coverage, but
  are not live-certified.
- Two current source rows are agentic and paid: `GET /v2/entities/get-brief` (`generateBrief`)
  and `GET /v2/entities/ask-entity` (`askEntity`). Neither was called. Their live cells are
  uncertified pending a captain decision; no browser session or alternate authentication was used.
