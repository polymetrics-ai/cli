## Task Delivery Header

- Issue: Refs #4283 — chore(connectors): pin and declare daily-use public API cohort
- Base branch: main
- Merges into: main
- Delivery: One PR against `main` containing sequential declaration-only increments of 10–20 connectors. This lane makes no live-certification claim and uses no provider credentials; each connector remains live-certification-pending.
- Working branch: fm/cli-top100-declaration-batch-r1
- Task: Pin only publicly retrievable provider API descriptions and derive connector-local, source-backed declarations. Exclude `github` and `zoom`; do not use provider credentials or change foundation code. Increment 1 targets Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.
- Verification: `scripts/gsd doctor`; `go run ./cmd/agentcontractgen check`; per-connector foundation checks; `connectorgen validate`; `surface-sync`; credential-free fixture/contract proofs; generated docs checks.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The planned source is public and reproducibly pinned | green | `SOURCE-LOCK-VERIFICATION.json` verifies byte count and SHA-256 for all 10 locally retained public downloads against their source locks. |
| Every declared operation has a corresponding API surface entry | green | `PROGRESS-LEDGER.json` records 4,378 / 4,378 source method/path bindings (100%) and `connectorgen validate` passes with zero findings. |
| Live certification remains pending without credentials | green | Generated `certification-sweep.json` artifacts byte-check successfully for all 10 bundles and the ledger explicitly records `live_certification: pending`. |

## Scope Interpretation

Captain clarification on 2026-08-19 defines this as a parity-declaration lane, not a certification lane: the one-connector certification rule does not constrain its declaration increments. This work makes no live certification claim, changes no engine/foundation code, and records every unavailable executor or unsafe/elevated operation as disabled with evidence. No files under `defs/github/` or `defs/zoom/` may change.

## Increment 1 Plan

1. Retrieve each provider's public machine-readable API description without credentials and pin source URL, retrieval time, SHA-256, byte count, method counts, and operation inventory under the connector's `sources/` directory.
2. Reconcile the pinned method/path inventory with the existing `api_surface.json`; retain every existing covered or blocked disposition and record any source/api-surface drift as an explicit rejection or foundation gap.
3. Materialize missing empty operation/write ledgers only when the bundle has no executable operation/action of that kind; never create a fake request or response schema. Declare source transport only when the existing streams have the registered declarative source executor; declare destination transport only when an eligible existing typed action and acknowledgement contract exist.
4. Update per-connector progress, rejection, and foundation-gap records; live certification remains `pending`, never passed or failed.
5. Run red/green artifact checks, connector validation, surface synchronization, source/api-surface inventory checks, conformance/fixture tests where present, generated docs checks, and the repository gates appropriate to these JSON changes. Commit the increment before reporting its measured elapsed time and file count.

## Foundation Check

| Need | Evidence required | Increment-1 disposition |
| --- | --- | --- |
| Declarative stream source | Registered `declarative_stream_source` executor and existing fixture-backed stream | Declare only where the bundle already has stream fixtures; validate structurally. |
| Reverse destination | Existing typed action plus durable acknowledgement and registered destination executor | Do not invent a transport. Record `foundation-gap` when only raw REST writes exist. |
| Live certification | Credentialed provider interaction, bounded cleanup/receipt, accepted artifact | Explicitly pending; prohibited by task scope. |

The concrete registration gap was confirmed for this cohort. No `sync_transport.json` was emitted because the generic source/destination adapter is not registered; `FOUNDATION-GAPS.md` links #4093 and exact engine/test evidence.

Transport gate follow-up: Captain review required the omission to be explicit before increment 2. `TRANSPORT-GAP.md` records the path-(b) decision, the runtime evidence, the smallest safe recovery, and ten recoverable rejection entries. No new descriptor may use GitHub's evidence or issue-label action contract for another connector.

## Increment 2 Plan

1. Retrieve and pin the confirmed public provider artifacts for Gitea, Grafana, Trello, Slack, n8n, Google Calendar, Gmail, Twilio, Amazon SQS, and Elasticsearch. Preserve the source artifact's actual document format: OpenAPI/Swagger where published, Google Discovery metadata for the Google APIs, and the AWS-owned SQS service model for the native Query API.
2. Mechanically reconcile every documented method/path with its bundle's `api_surface.json`. Preserve an existing enabled binding only when it is the same documented method/path; add every remaining source operation as a disabled declaration with the fixed-vocabulary reason and source evidence. Do not create request, response, pagination, or body schemas.
3. Retain existing typed operations and writes, add empty ledgers only where they are absent, and ensure every source DELETE is either represented by an existing delete action or explicitly disabled. Do not manufacture `sync_transport.json`: add a recoverable `foundation-gap` record referencing #4093 for both transport directions unless the declared adapter, evidence constants, action binding, and acknowledgement contract all already exist.
4. Extend the source-lock verification, progress ledger, rejection list, foundation-gap log, TDD/verification evidence, connector certification sweeps, and generated documentation evidence. No live provider calls are permitted; live certification stays `pending`.
5. Run the declaration and generated-artifact gates for the ten affected bundles, then commit the increment and report its elapsed time and file count before starting increment 3.

## Lifecycle Record

- Inline/manual GSD fallback: this execution environment does not provide the compatible Pi runtime and the canonical contract forbids spawning GSD roles. Generated and reviewed: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, `code-review` for issue `4283`.
- Skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- CLI parity reference reviewed. It becomes applicable after a connector-local declaration changes a generated command surface.

## Captain full-parity correction — Docker Hub first

Captain's 2026-08-19 work order suspends Increment 2 after its draft inventory
was preserved on `fm/cli-top100-declaration-batch-r1-inc2-wip`. Before any new
connector, the ten Increment-1 source locks must each account for every pinned
operation with a source-backed typed contract or an itemized disabled
disposition. Docker Hub is the first proof slice.

### Docker Hub source facts and intended disposition

The pinned Docker Hub OpenAPI 3.0.3 artifact contains 54 REST operations: 24
GET, 12 POST, 5 PATCH, 4 PUT, 6 DELETE, and 3 HEAD. Its existing v2 API
surface maps all 54 exactly once: four GET routes already back declared ETL
streams, 47 routes are authenticated/elevated-scope blocks, and three HEAD
routes have no response-body executor. `POST /v2/users/login` is explicitly
deprecated in the pinned source. The plan is therefore to derive connector-local
source-contract inventory entries for the 23 JSON GET and 26 mutating
non-deprecated routes, retain the four stream bindings, and record all 54 rows
in an immutable crosswalk/disposition artifact. The three HEAD operations and
deprecated login remain disabled with their fixed-vocabulary reasons.

No Docker Hub source-contract inventory entry is a terminal direct-read or
direct-write command. A terminal command would additionally require its own
complete request/response/pagination or body contract, output policy, fixture,
and Foundation Check. This slice intentionally leaves the existing blocked API
surface rows blocked, records live certification as pending, and does not create
`sync_transport.json` while #4093 remains open.

### Docker Hub Foundation Check

| Need | Evidence | Disposition |
| --- | --- | --- |
| Source-operation contract inventory | Pinned OpenAPI method/path/parameters/content-type plus source location | Materialize `operations.json` only for source-supported GET and mutating methods; each row crosswalks to its exact API-surface method/path. |
| Direct-read/direct-write terminal command | Complete source contract, fixture, command surface, and real runtime preflight | Not claimed. Inventory rows remain non-terminal and their API-surface row remains blocked. |
| HEAD operation | A bounded response-less status/existence executor | `foundation-gap`; retain an explicit disabled disposition. |
| Deprecated login | Pinned source marks `deprecated: true` | `provider-does-not-expose`; do not materialize a terminal contract. |
| Sync transport | Connector-neutral registration, connector evidence, and destination acknowledgement | Existing recoverable #4093 record; no placeholder descriptor. |

## Captain correction — elevated scope is runtime authorization, not a disabled declaration

Captain's 2026-08-19 correction supersedes the Docker Hub elevated-scope
disposition above. A source operation which merely requires an administrator,
organization role, SCIM permission, or bearer token remains **enabled as a
source-contract declaration**: the required permission is retained as
source-backed security metadata and an actual `403` is a runtime authorization
outcome. It must not be counted as disabled or entered in the rejection list.

The exception is an operation whose purpose is to mint, exchange, list, rotate,
or revoke credentials/session tokens. Docker Hub's documented login, two-factor
login, auth-token, personal access-token, and organization access-token routes
remain disabled as `unsafe-to-exercise`; their source contract remains in the
inventory so the omission is never hidden. The three HEAD operations remain a
recoverable `foundation-gap`. The members CSV export is `schema-incompatible`:
the pinned document exposes `text/csv`, but the connector declaration has no
source-backed bounded byte contract for an executable binary transfer.

This correction does not turn a source-contract inventory record into a
terminal CLI command. Terminal direct read/write still requires the separate
complete command contract and preflight evidence. Docker Hub reporting must
therefore distinguish 100% declared coverage from the enabled source-contract
percentage, and state that live certification remains pending.

## Captain deliverable correction — runnable command/action parity

Captain's 2026-08-19 deliverable correction supersedes the preceding
inventory-only stopping condition. An operation contract which no
`cli_surface.json` command binds is not user capability: `surface-sync` only
reconciles metadata for commands already declared. Docker Hub is not complete
until each executable pinned operation has a typed contract, a runnable command
whose real preflight reaches the expected credential/lifecycle boundary, and
each create/update/delete has a source-backed `writes.json` action (including
every safe ordinary delete). Existing ETL streams satisfy their four mapped
read routes; remaining executable GETs require `direct_read` commands and
mutations require `reverse_etl` commands plus typed write actions so the shared
plan, preview, approval, and execute boundary remains in force.

The Docker Hub target is therefore 33 executable documented operations: four
existing ETL commands, 13 new direct reads, and 16 new reverse-ETL commands
with 16 write actions, four of them deletes. The remaining 21 rows are the only
disabled set: 13 credential/session routes (`unsafe-to-exercise`), five
executor/pagination routes (`foundation-gap`), and three unsupported
source-content routes (`schema-incompatible`). The pinned OpenAPI is the sole
source for parameter, request-body, response, and pagination declarations;
unknown constraints stay disabled rather than guessed.

## Captain correction — secret risk is a foundation gap, not unsafe refusal

Captain's later 2026-08-19 clarification supersedes the target and disposition
in the preceding paragraph. Docker Hub's eight personal/organization
access-token list, detail, update, and delete routes return metadata, not the
secret token: they are runnable via four direct reads and four typed
reverse-ETL commands. The target is therefore 41 executable operations: four
ETL, 17 direct reads, and 20 reverse-ETL commands with six typed deletes.

The two token-create responses expose `token`; login and 2FA return `token`
(and login's continuation can return `login_2fa_token`); auth-token returns
`access_token`. Those exact source fields are marked in the operation
dispositions, the affected operation contracts are `secret_sensitive` with
`sensitive_policy`, and exact secret request fields are `x-secret` in
`spec.json`. They remain declared, disabled, recoverable `foundation-gap`
because `internal/connectors/engine/bundle.go:2772-2776` says live secret
writes are not implemented, and a typed secret-response storage/redaction
contract is also absent. Docker Hub's `unsafe-to-exercise` count is zero.

For every later connector, use this criterion: only a live operation that is
genuinely destructive or irreversible without user intent belongs in
`unsafe-to-exercise`; a secret ingress/egress limitation is a named,
recoverable foundation gap.

## Captain complete-map order — before certification

### Red

The source locks and inventories alone did not prove a stable certification
target: the ten bundles did not consistently show the user-facing CLI binding,
binary route, transport admission, or exact foundation gap for every pinned
method/path. The old GitHub-shaped sync descriptor was also not a valid basis
for a new connector declaration.

### Green

Docker Hub `3ee815c01` is retained as the accepted reference. The other nine
bundles receive a source-lock crosswalk and declaration-disposition map in
`internal/connectors/defs/<connector>/sources/`. Each row names exactly one of
direct read, direct write, ETL, reverse ETL, binary read, or binary write and
records the present CLI foundation or a named recoverable gap with file/line
evidence and minimal change. `COMPLETE-PARITY-MAP.md` records the batch
denominators, command-derived enabled percentages, deletes and gap IDs.

ETL and reverse ETL are assessed only using the definition-owned transport
contract in `docs/sync-transport-definition.md` (PR #4286). Every connector is
an explicit two-direction transport declaration-pending record until it provides
its own exact executor, delivery facts, typed action strategy/acknowledgement,
and conformance evidence. No `sync_transport.json` is copied or invented.

### Refactor

Do not run certification until this map is stable. This definitions-only change
does not add engine code, provider schemas, credentials, or live provider I/O.
