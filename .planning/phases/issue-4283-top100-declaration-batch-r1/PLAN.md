## Task Delivery Header

- Issue: Refs #4283 — chore(connectors): pin and declare daily-use public API cohort
- Base branch: main
- Merges into: main
- Delivery: Existing PR #4294 is retargeted to `main` after a non-rewriting merge of current `origin/main`; it is complete only after all ten in-scope bundles have a source-backed, installed-binary-reachable command disposition for every documented operation and each eligible typed action has a definition-owned reverse-ETL destination. No provider credential or live request is permitted; live certification remains pending.
- Working branch: fm/cli-top100-declaration-batch-r1
- Task: Reconcile only Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana after the #4303 typed-destination foundation. Preserve their pinned public artifacts and derive every command/transport declaration from them. Exclude `github`, `zoom`, all other connectors, shared runtime changes, credentials, and live provider I/O.
- Verification: `scripts/gsd doctor`; `go run ./cmd/agentcontractgen check`; targeted red/green integrity assertions; per-connector `connectorgen validate`, `surface-sync --check`, and certification-sweep checks; fixture/conformance and commandrunner preflight; generated manual/website checks; `make connector-runtime-preflight`, `make connector-canon-check`, `go run ./cmd/connectorgen boundary . --json`, and `make verify`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The planned source is public and reproducibly pinned | green | `SOURCE-LOCK-VERIFICATION.json` verifies byte count and SHA-256 for all 10 locally retained public downloads against their source locks. |
| Every found source operation has a corresponding API surface entry | green | `PROGRESS-LEDGER.json` records 4,378 operations found and 4,378 source method/path bindings, each with high confidence based on a complete provider-published OpenAPI source; `connectorgen validate` passes with zero findings. |
| Live certification remains pending without credentials | green | Generated `certification-sweep.json` artifacts byte-check successfully for all 10 bundles and the ledger explicitly records `live_certification: pending`. |

## Scope Interpretation

Captain clarification on 2026-08-19 defines this as a parity-declaration lane, not a certification lane: the one-connector certification rule does not constrain its declaration increments. This work makes no live certification claim, changes no engine/foundation code, and records every unavailable executor or unsafe/elevated operation as disabled with evidence. No files under `defs/github/` or `defs/zoom/` may change.

## Current reachability supersession — 2026-08-20

The later captain decision supersedes the preceding historical disabled-command
language: every documented operation remains a target for installed-CLI
execution through its exact, source-locked typed declaration. The authoritative
classification is `EXECUTABLE-OPERATION-CAPABILITY-AUDIT.json`; neither safety,
privilege, tier, destructive behavior, nor the absence of live credentials is
an active reachability exclusion. The connector lane still makes no shared
runtime edits; its job is to dispatch the closed foundation slices recorded
below and to materialize their connector-owned declarations.

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

PR #4286 subsequently supplied the connector-neutral declarative source
factory. The ten `sync_transport.json` source declarations now carry concrete
stream allowlists and per-connector evidence. The source-only production route
is structurally proved by App open/composition; it is not live certification.
The only remaining transport claim is the truthful reverse-ETL foundation gap
`generic-typed-destination-executor`, documented in `TRANSPORT-GAP.md`.

## Increment 2 Plan

1. Retrieve and pin the confirmed public provider artifacts for Gitea, Grafana, Trello, Slack, n8n, Google Calendar, Gmail, Twilio, Amazon SQS, and Elasticsearch. Preserve the source artifact's actual document format: OpenAPI/Swagger where published, Google Discovery metadata for the Google APIs, and the AWS-owned SQS service model for the native Query API.
2. Mechanically reconcile every documented method/path with its bundle's `api_surface.json`. Preserve an existing enabled binding only when it is the same documented method/path; add every remaining source operation as a disabled declaration with the fixed-vocabulary reason and source evidence. Do not create request, response, pagination, or body schemas.
3. Retain existing typed operations and writes, add empty ledgers only where they are absent, and ensure every source DELETE is either represented by an existing delete action or explicitly disabled. Declare the reusable ETL source only when the bundle has concrete streams and source evidence. Do not invent destination `transport_binding` actions: record `generic-typed-destination-executor` until a definition-selected typed destination factory exists.
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
surface rows blocked and records live certification as pending. The later PR
#4286 source declaration is governed by the current transport correction below.

### Docker Hub Foundation Check

| Need | Evidence | Disposition |
| --- | --- | --- |
| Source-operation contract inventory | Pinned OpenAPI method/path/parameters/content-type plus source location | Materialize `operations.json` only for source-supported GET and mutating methods; each row crosswalks to its exact API-surface method/path. |
| Direct-read/direct-write terminal command | Complete source contract, fixture, command surface, and real runtime preflight | Not claimed. Inventory rows remain non-terminal and their API-surface row remains blocked. |
| HEAD operation | A bounded response-less status/existence executor | `foundation-gap`; retain an explicit disabled disposition. |
| Deprecated login | Pinned source marks `deprecated: true` | `provider-does-not-expose`; do not materialize a terminal contract. |
| Sync transport | Registered source executor, connector evidence, and typed destination acknowledgement | Declare the source-only contract; keep reverse ETL as `generic-typed-destination-executor` until its typed destination exists. |

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
therefore report operations found, source-input confidence and mapped rows
separately from the enabled source-contract percentage, and state that live
certification remains pending.

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
contract in `docs/sync-transport-definition.md` (PR #4286). Every connector
now declares the registered source role with connector-owned evidence. Reverse
ETL is one recoverable `generic-typed-destination-executor` foundation gap: no
destination binding, acknowledgement, strategy, or generic HTTP writer is
invented.

### Refactor

Do not run certification until this map is stable. This definitions-only change
does not add engine code, provider schemas, credentials, or live provider I/O.

## Captain classification correction — direct write versus reverse-ETL eligibility

The earlier six-primary-class wording is superseded. A documented mutation is
`direct_write`, including a typed create, update, upsert, or delete action.
Reverse ETL is an attribute on that direct-write operation, not an endpoint
class and not an inference from a typed action. It is eligible only after the
definition carries a destination transport binding, per-mode apply strategies,
and an acknowledgement contract.

For all ten bundles the `reverse_etl_eligibility` attribute is currently
`foundation-gap` / `generic-typed-destination-executor`: the only production
destination factory remains the issue-label-specific one at
`internal/app/issue_label_warehouse_transport.go:85-95`. The map must report
five primary endpoint classes plus this separate eligibility. It must report
enabled direct writes as direct writes, while reverse-ETL eligibility remains
zero until the connector-neutral typed destination factory lands.

## Source-lock completeness correction

The Batch-1 inputs pass the captain’s source-lock completeness check: every
lock records `counts.total`, per-method counts and an equal-sized operation
inventory from a complete provider-published OpenAPI document. The maps report
`operations_found` and high-confidence machine-readable-spec basis instead of
a self-referential declaration percentage. PR #4294 remains held while the
fleet-wide suspect-lock correction is completed.

## CI verify gap — shared source-evidence assertion

### Red

The pushed PR verification run `32271368383` fails
`TestDefinitionTransportFactoriesSelectDeclaredEvidence`. The test assumes
GitHub's conformance record occupies `DefinitionFactory.SourceEvidence`, but
the shared declarative source factory intentionally selects the first
declaration's evidence and admits the rest through `AcceptedSourceEvidences`.
Batch-1 declarations make Asana first, while GitHub's evidence is correctly
present in the accepted set.

### Green

The test must assert the public factory contract: GitHub's declared evidence
is accepted by the shared source factory, whether it is the primary evidence
or one of its additional accepted references. It must not encode registry
iteration order. The destination assertion remains exact because it has one
definition-owned destination declaration.

### Refactor

Change only the failing test assertion and its delivery evidence. Do not alter
the source-factory selection, connector declarations, or conformance data to
make GitHub happen to sort first. Re-run the exact failing test before the
affected package and static declaration gates.

## Authorized downstream shared-test repair — 2026-08-24

### Decision and scope

Firstmate authorized the two shared downstream test repairs after reviewing
the fixed-cohort evidence. This is an inline/manual GSD execution because this
environment has no compatible isolated Pi worker and the canonical contract
forbids role spawning. The task-delivery header at this document's start
continues to govern PR #4294 (`fm/cli-top100-declaration-batch-r1` to `main`).

Only these files may change for this slice:

- `cmd/connectorgen/operationevidence_test.go`;
- `internal/app/transport_composition_test.go`; and
- this issue's GSD/TDD/verification evidence.

The immutable `internal/connectors/operation-evidence-fixed-100.json` must
not be regenerated, re-selected, or rewritten. No production source factory,
connector declaration, source lock, generated artifact, or provider contract
may change. Skills: `golang-how-to` and `golang-testing`.

### Red

`TestOperationEvidenceFixed100RejectsEveryRegression` and
`TestOperationEvidenceCheckRunsFixed100Gate` currently load the 100-row
cross-connector cohort in a temporary repository that contains only GitHub;
they fail at `asana.rest.getCustomFieldsForWorkspace`. Separately,
`TestDefinitionTransportFactoriesSelectDeclaredEvidence` assumes GitHub's
conformance record is the primary evidence of a registry-wide shared factory,
but Asana is the deterministic first declaration and GitHub is correctly
present in the accepted evidence set. The exact failed outputs and expected
versus produced values are preserved in
`VERIFY-SHARED-TEST-FIXTURE-EVIDENCE-2026-08-24.md`.

### Green plan

1. Read the existing fixed reference inside the temporary workspace, derive
   the connector prefix of every `source_id`, and copy exactly those definition
   trees plus their generated website rows. Do not change the fixed reference.
2. Retain the fixed-reference validator and add a source-row-removal proof:
   remove one GitHub source row only in the temporary test workspace and show
   both the direct validator and the CLI `--check` gate refuse it.
3. Keep the destination assertion exact. Replace only the source-factory
   primary-order assertion with an exact-membership assertion over the
   factory's primary and accepted source-evidence records, requiring GitHub's
   declared record specifically.
4. Run the two targeted tests red/green, their package tests, then the
   repository declaration/generated gates and `make verify` required by the
   delivery contract. Record every result before a commit.

### Green result

The fixed reference remains unchanged. The workspace now derives its six
definition trees and generated website rows from the fixed reference itself;
the exact source set is Asana, Bitbucket, CircleCI, Docker Hub, GitHub, and
Jira. Both fixed-cohort test paths deliberately remove
`github.rest.issues/list-for-repo` only from a `t.TempDir` source lock and
assert the source-specific failure: the direct validator and the CLI `--check`
gate each reject the missing operation. The app test now requires GitHub's
exact conformance reference in the shared factory's primary-or-accepted set,
while retaining its exact destination evidence assertion.

Focused results:

```text
ok  polymetrics.ai/cmd/connectorgen  16.514s
ok  polymetrics.ai/internal/app      2.896s
```

### Refactor boundary

No assertion is deleted or broadened to accept arbitrary evidence. The fixed
cohort's prior 100 GitHub anchors remain validated in operation evidence even
though the current selection fixture itself contains non-GitHub rows. The
transport assertion remains evidence-specific and does not depend on registry
iteration order.

## PR #4297 repair loop — observed Docker Hub outcome

### Red

Docker Hub reported 41 of 54 enabled before the repair. The proof must traverse
the full bundle loader and command surface, rather than trusting an executor
unit test alone.

### Green

PR #4297 enabled the two source-derived paginators and two closed-SCIM JSON
direct writes. `connectorgen validate`, `surface-sync --check`, sweep
generation, and four real no-credential preflight calls agree that Docker Hub
is now 45 of 54 enabled. The SCIM update's source `allOf` object plus
description is mechanically normalized to its exact object-plus-annotation
shape for the engine's closed schema dialect; no field is invented.

The other four intended flips do not load: the `rest_status` and `text_export`
executors exist, but the operation-kind loader/validation path at
`internal/connectors/engine/bundle.go:2451,2676,2705,2733` omits both kinds.
The rows remain a recoverable shared `operation-kind-loader-registration`
foundation gap. The required foundation change is to register `rest_status` and
`text_export` in the bundle loader operation-kind switch and validation so a
definition can declare them. This definitions-only shard does not alter engine
code.

### Refactor

Remove the four resolved rejection records and the retired pagination and SCIM
media-type gap IDs. Retain the five secret-response pending rows,
`generic-typed-destination-executor`, all source-lock counts, and the
live-certification-pending boundary. Do not misclassify the missing operation
block mapping as a Docker Hub or provider-schema limitation.

## Local verify recovery — Notion operation ledger

The full repository gate exposes one stale Batch-1 audit representation:
`POST /v1/search` already has two source-qualified response-arm rows and must
not also have an unqualified third row. Preserve the source operation through
those two rows, restore `named_dependency=sensitive_policy` to the three
recoverable OAuth rows, and prove the exact ledger test before repeating
`make verify`. This is an audit-artifact repair, not a change to a provider
operation, runtime executor, or reverse-ETL declaration.

## Reverse-ETL preparation — Docker Hub typed actions

Before #4303 makes destinations declarable, keep the source-backed action
inventory explicit: 20 of Docker Hub's 27 direct writes already bind exact
typed actions; five credential lifecycle/session operations are not sync
targets; and two SCIM writes require a typed-action request-content-type
extension because `writes.json` currently emits `application/json`. Record the
two-operation `typed-action-content-type` foundation gap with source and engine
evidence, and do not fabricate a destination or `transport_binding` entry.

## Reconciliation relaunch — 2026-08-20 (supersedes prior destination-gap completion claims)

PR #4304 (`fm/cli-reverse-etl-destination-r1`) is now merged into this branch
as commit `192180675`. PR #4294 is API-retargeted to that exact branch and
must retain this stack; it does not target `main` directly. The scope is fixed
to the ten existing Batch-1 connector directories: `dockerhub`, `notion`,
`stripe`, `bitbucket`, `gitlab`, `circleci`, `sentry`, `vercel`, `asana`, and
`jira`.

### Completion contract

For every source-lock operation, reconcile the seven surfaces: `binary_read`,
`binary_write`, `direct_read`, `direct_write`, `etl`, `reverse_etl`, and its
actual `pm <connector> ...` command. Every documented operation remains in the
source map and has either a real installed-binary command or a precise engine
incapability with the refusing file, line, and minimal missing hook. Privilege,
destruction, expense, rarity, or lack of live credentials is not a reason to
omit or disable a representable operation; safety metadata and the existing
plan → preview → approval → execute boundary govern mutations.

Every eligible typed `writes.json` action must be matched by an exact
`sync_transport.json` destination declaration: definition-owned action binding,
allowed input fields, per-mode apply strategies, acknowledgement/delivery
facts, conformance evidence, and no tombstone claim unless its source-backed
contract proves it. Binary transfer remains a distinct binary surface and is
never relabelled as a REST write or reverse destination.

### TDD reconciliation slices

1. **Red — surface audit.** A read-only seven-surface audit of the ten source
   maps must fail while a documented endpoint has no command or an eligible
   typed action has no exact destination binding. It records individual rows,
   not an aggregate availability percentage.
2. **Green — connector-local declarations.** In bounded connector commits,
   derive only supported declarations from the pinned source evidence; run the
   matching focused generator, conformance, preflight, and generated-document
   checks. A represented destructive or privileged mutation remains reachable
   through its canonical command and the existing approval boundary.
3. **Refactor/evidence.** Regenerate connector-owned artifacts only, update
   the machine-readable seven-surface ledger and human summary, retain every
   live-provider check as `pending`, and record every genuine engine refusal in
   the existing foundation-gap log. Never add a generic writer or edit shared
   engine code in this connector lane.

### Required parity checklist

- [ ] Source lock, operation crosswalk, and declaration disposition still
  account for every provider operation in each of the ten bundles.
- [ ] Every eligible typed action has an exact destination declaration and
  reverse-ETL remains warehouse-mediated and approval-gated.
- [ ] Every source operation has an installed-binary command or an exact,
  source-backed engine refusal; no safety/premium/scope-only omission remains.
- [ ] `sync_transport.json` distinguishes binary from REST and supplies exact
  source/destination evidence, input fields, strategies, acknowledgement, and
  delivery facts.
- [ ] Generated manuals/catalog/website data and runtime command help are
  regenerated or explicitly verified unchanged for every affected connector.
- [ ] Live provider certification remains pending without credentials; fixture,
  structural, and generated-artifact checks are recorded separately.

### Lifecycle and ownership

- Inline/manual GSD fallback: prompts for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review` were resolved through
  `scripts/gsd`; compatible isolated Pi workers are unavailable and the
  single-worker connector contract forbids role spawning.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation`.
- CLI help/manual/website parity is in scope because the task changes connector
  command surfaces. The project handoff, current connector canon,
  implementation procedure, remote reproducibility guidance, migration
  conventions, architecture v2, and certification design were reviewed before
  declaration edits.

## Executable-operation foundation plan — 2026-08-20

Captain rejected a disabled-command placeholder: a source operation is
reachable only when an installed CLI command executes its exact typed provider
contract. `EXECUTABLE-OPERATION-CAPABILITY-AUDIT.json` reclassifies all 3,366
unbound source operations by source identity and next capability: 1,389 fixed
REST reads, 1,828 fixed REST writes, 120 bounded binary transfers, 10 status
registrations, and 19 provider schemas too open to represent without a raw
escape hatch.

The implementation plan is `EXECUTABLE-OPERATION-FOUNDATION-DESIGN.md`:
rehydrate only hash-matched public source artifacts, materialize fixed operation
contracts, add closed typed headers, compose #4305 structured bodies, add
bounded binary/status/text paths, then prove #4304 App/CLI destination dispatch
and action-scoped source mapping. No caller-selected URL, verb, header, body,
arbitrary JSON, or generic transport is acceptable.

## Captain hard pre-merge gate — 2026-08-20

The `SEVEN-SURFACE-RECONCILIATION.json#pre_merge_gate` is the authoritative
merge gate and currently reports `merge_ready: false`. It requires a
source-locked, provider-documentation two-way trace for every operation through
the canonical ledger, runtime declaration, generated CLI command, and generated
website representation. ETL, reverse ETL, direct read, direct write, binary
download, and binary upload are separate semantic surfaces. An `N/A`
disposition requires provider evidence that the surface is semantically absent;
it cannot be justified by privilege, cost, destructive effect, safety concern,
or missing credentials.

Each applicable operation must retain fixture/conformance and output-preserving
runtime evidence plus exact generated CLI/help/manual/website drift proof. Zoom,
Twenty, and Gong require authorized provider-live certification in their own
lanes. This PR remains non-live and must not claim any provider-live result
without credentials and accepted evidence. Implementation remains paused until
F0, F2/F4, #4305, and the final #4304 heads publish.

The planning record is necessary but insufficient. The integration branch must
then implement a CI-suitable executable repository validator for the fixed 100
with schema-backed per-connector and aggregate verdicts. Its negative suite
must fail for a missing source hash; missing or unreachable command; CLI/website
drift; missing runtime fixture/conformance evidence; surface misclassification;
binary download/upload collapse; a disabled callable operation; and omission of
non-secret output. Only a passing validator result may satisfy this final gate.

## Missing-foundation mapping deliverable — 2026-08-20

`MISSING-FOUNDATION-DELIVERABLE.json` is the authoritative deduplicated
foundation-gap mapping for the current ten-connector batch. Its eight stable
gap definitions own capability, evidence, issue/lane, exact closure rule, and
cross-connector fanout; operation/action membership rows supply the exact
provider operation plus source URL/revision/hash and affected surface. It
records 3,366 source-operation and 491 typed-action rows as open and not
enabled. The 28 actions without a resolvable current source identity are
explicit F0/F1 gaps, never inferred bindings.

No open shared gap is a connector-specific workaround, `disabled`, or `N/A`,
and none contributes merge-ready credit. The rollup is intentionally bounded:
ten mapped connectors of the fixed 100, with the remaining 90 marked
unmapped rather than fabricated. F2 has a zero-count pending fanout until F0
imports exact required headers.

## Main-base reconciliation — 2026-08-23

Captain has replaced the temporary stacked-base delivery with a main-base
delivery. The reconciliation uses `git merge origin/main` (never a rebase or
force-push); `origin/main` is pinned at `6410fe59c`, which contains the
0.3.0 foundation rollup and declaration-owned operation routes as squash
content. Conflicts retain current main for foundation/route changes and this
lane only for the ten Batch-1 connector declarations and their evidence.

### Reconciliation red/green evidence

- **Red:** PR #4294 targets the stale `fm/cli-reverse-etl-destination-r1`
  base, so even its green checks cannot land the Batch-1 declaration cohort in
  `main`.
- **Green:** merge `origin/main`, retarget #4294 through the GitHub REST API,
  then rerun the full repository `make verify` gate. Recompute the
  source-lock ledger and assert exactly 4,378 documented operations found and
  exactly 4,378 declared across Docker Hub, Notion, Stripe, Bitbucket, GitLab,
  CircleCI, Sentry, Vercel, Asana, and Jira.
- **Refactor:** retain only this reconciliation evidence; no source operation,
  foundation capability, connector directory, credential, or live-provider
  certification scope changes as part of moving the base.

The GSD prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` were re-resolved with `scripts/gsd` for this
amendment. The execution remains inline because this environment has no
compatible Pi runtime and the canonical connector contract forbids role
spawning. The reconciliation reuses the existing focused red/green checks;
its terminal verification is the full `make verify` gate plus API read-back of
PR #4294's `main` base.

## Generated skills synchronization and Docker Hub preflight comparison — 2026-08-24

### Decision and scope

Firstmate classified `docs/skills` drift as an in-scope downstream artifact
repair because the declaration batch changes the generated connector surface.
Only generated `docs/skills/**` and delivery evidence may change. The two
Docker Hub SCIM preflight failures are not assumed pre-existing: compare the
same runtime-preflight test against a clean current `origin/main` tree before
changing any SCIM declaration. Required skills reviewed: `golang-how-to`,
`golang-testing`, and `golang-documentation`; `pm help skills` confirms the
generator is metadata-only and credential-free.

### Red

`GOFLAGS=-p=3 make verify` reports tracked `docs/skills` drift and reports
Docker Hub `scim user create`/`scim user update` preflight failures caused by
source-derived request-schema `example` keywords.

### Green plan

1. Run `pm skills generate --dir docs/skills`, review the diff, and run the
   exact generated-skills parity test.
2. Export a disposable clean current-main tree inside this worktree, run the
   same Docker Hub preflight test, compare it with this branch, and remove only
   the agent-created temporary tree.
3. If main fails identically, record a proven external finding and leave SCIM
   untouched. Otherwise repair the branch regression before continuing.
4. Re-run required local checks and obtain the PR CI/check plus mergeability
   rollup before any delivery action.

### Refactor boundary

Do not regenerate source locks, reselect the fixed cohort, change shared
runtime behavior, or use credentials. A working command must not be moved to
partial merely to obtain a green preflight result. If the pinned source
faithfully describes a request body the installed executor refuses, record the
source-traced engine gap and test any candidate disposition against immutable
execution evidence; never invent a closed body schema to preserve an
implemented claim.

### Green result / required branch repair

`GOFLAGS=-p=3 go run ./cmd/pm skills generate --dir docs/skills --json`
changed exactly the ten Batch-1 connector skill files. Its parity gate passes:
`GOFLAGS=-p=3 go test -timeout 20m -count=1 -run
'^TestSkillsGenerateMatchesTrackedSkills$' ./internal/cli`.

The exact preflight comparison disproves the prior pre-existing assumption:

```text
origin/main 3c394a0e: PASS  TestEveryImplementedCommandPassesRuntimePreflight
branch 2e860dfe3:  FAIL  dockerhub scim user create/update
                         unknown keyword "example" in source-derived body schemas
```

The branch must therefore merge current `origin/main` (never rebase) and then
rerun the preflight. The temporary clean-main archive was moved to Trash after
the check; no provider request or credential was used.

### Post-merge red and source-faithful correction

The non-rewriting merge at `f528b806d` removes the `example` keyword refusal,
but the same structural preflight then rejects the two SCIM operations because
the pinned source has open object schemas. The public artifact hashes to the
locked `99d9d53c...53d0756`; lines 3921-3959 define `scim_user_name` and
`scim_user` with `properties` and no `additionalProperties: false`. The
executor refuses that exact source form at
`internal/connectors/engine/structured_rest_body.go:1436-1444` before any
request can occur. Adding the missing keyword to `operations.json` would
invent a provider constraint.

An attempted source-faithful `partial` disposition was rejected by the
immutable fixed-100 evidence gate: `connectorgen operation-evidence --check`
reports `dockerhub.rest.post_/v2/scim/2.0/Users execution evidence regressed`.
The current command declarations are restored unchanged, because a partial
placeholder would violate the complete-reachability contract just as surely as
an invented closure would violate the pinned-source contract. The exact
recovery is a typed bounded representation for source-declared open objects
without a raw/generic HTTP writer.
