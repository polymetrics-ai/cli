# Issue 4325 — Batch 1 Connector Repair Plan

## Task Delivery Header

- Issue: Refs #4325 — restore batch 1 independent gate
- Base branch: main
- Merges into: main
- Delivery: prepare connector-owned mapping commits on
  `fm/cli-batch1-repair-r1`; do not open a pull request or merge while the
  shared validator review is outstanding. The eventual pull request targets
  `main` and requires the independent credential-free Gate B rerun returning
  GO and all repository checks green.
- Working branch: fm/cli-batch1-repair-r1
- Task: Repair the ten batch-1 source bundles and their derived command,
  operation, and evidence surfaces without credentials or reduced quality gates.
  Each connector is a sequential ownership slice; any missing shared runtime
  capability is split into a foundation issue instead of being shimmed locally.
- Verification: source-import/check, `connectorgen validate`, `surface-sync
  --check`, real commandrunner preflight, focused regression tests, built-binary
  credential-boundary probes, GitHub source-lock checksum assertions, full
  `make verify`, and an independent Gate B rerun.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Current provider locks map exactly to declared operation identities | live | A fresh credential-free source import accepts the pinned bytes and its descriptor identity set matches the provider document; a stale or missing operation fails the check. |
| Enabled command surface is executable | live | Real `commandrunner.Preflight` accepts every implemented command and a built binary dispatches selected commands to `missing --credential`, not `unknown command`. |
| Disabled evidence is truthful | live | The independent ledger audit finds zero `requires-elevated-scope` reasons and every remaining `foundation-gap` contains a current refusing `file:line`. |
| GitHub parity remains immutable | live | The GitHub lock and descriptor file byte counts and SHA-256 equal the captain-specified values. |
| No provider is falsely certified | live | `pm connectors inspect <connector> --json` retains `live_certification: pending` for every batch connector. |

## Required skills and workflow

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation`.
- GSD prompts resolved and executed inline: `discuss-phase 4325`,
  `plan-phase 4325 --tdd`, and `execute-phase 4325`. `verify-work` and
  `code-review` remain pending the complete Batch 1 gate. Inline execution is
  required because compatible isolated runtime agents are unavailable and the
  canonical contract forbids role spawning.
- CLI parity: this changes generated connector command surfaces. Verify the
  bare `pm connectors` namespace, `pm connectors inspect <name> --json`, one
  changed connector command `--help`, generated docs checks, and discovery
  metadata. Hand-authored CLI/website pages are not applicable unless a
  generator reports changed output.

## Gap plan — cited non-executable mutations (2026-08-24)

## Captain declaration-admission override (2026-08-25)

Provider-backed declaration completeness is an independent admission rule. Every
method/path operation in a retained provider source—read, binary, mutation,
delete, or reverse-ETL-shaped—must retain a source-cited JSON declaration even
when it is unsafe or its runtime foundation is absent. A non-runnable operation
is explicitly `deferred`/`foundation-gap` with its exact source citation and
refusing implementation location; it is never omitted, lowered to a different
operation class, or labelled `implemented`.

This creates two deliberately separate gates:

1. **Source declaration admission** proves a one-to-one source identity,
   canonical endpoint, classification, destructive metadata, and either an
   executable contract or an explicit deferred foundation reference.
2. **Runtime usability** retains the existing `availability: implemented` →
   real `commandrunner.Preflight` → credential-bound built-binary proof. It
   must continue to reject unsafe, incomplete, or unbound execution claims.

The proposed certification change is documented with its red/green matrix in
`TDD-LEDGER.md` and its present-artifact inventory in
`evidence/BATCH1-POST-060BB-REPORT-2026-08-24.md`. No certification code is
started until that design has been accepted as a separate implementation slice.

Captain direction is to retain every provider operation and mark a mutation
whose source-cited action cannot yet be executed as declared-and-deferred. The
only permitted mechanism is
`sourceProjectionApplyNonExecutableMutationDispositions` in
`cmd/connectorgen/sourceprojection.go`. Its fixed foundation tag is
`source-cited-non-executable-mutation-foundation-r1`; the generated descriptor
copies the provider URL, SHA-256, byte count, and document location from the
source operation and marks the gap merge-blocked.

- Active target: Sentry, then Asana and Jira. Each connector remains a
  sequential slice. Vercel is discovery-only until its absent CLI surface has
  been separately reported to the captain.
- Source-of-truth classification: retain the existing connector-owned
  `sources/<connector>-declaration-disposition.json` entry for every affected
  source ID. It supplies the audited `etl`, `reverse_etl`, `direct_read`,
  `direct_write`, `binary_upload`, or `binary_download` classification and
  provider trace; the mutation-disposition file is deliberately narrow and
  only adds the executable-action foundation gap.
- Hard boundary: never add a disposition for any method/path with a working
  `availability: implemented` command. The importer and validator both refuse
  that downgrade. Non-mutating reads without an executable path are reported
  as evidence gaps; no mutation-only workaround is invented.
- Manual GSD fallback: the canonical prompts were resolved with
  `scripts/gsd prompt`; compatible isolated runtime agents are unavailable and
  project policy forbids role spawning, so this gap is planned and executed
  inline. Required skills used: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, and `golang-testing`.

### Red / green contract

- **Red:** On 2026-08-24, `go run ./cmd/connectorgen source-import sentry
  --check` passed with 223 operations, while `go run ./cmd/connectorgen
  validate internal/connectors/defs/sentry` exited 1 with 32 exact
  `source operation has no executable action` findings.
- **Green:** after a connector-owned, strict v1
  `<connector>-mutation-dispositions.json` names each exact source ID, method,
  and path, `source-import --check` and `connectorgen validate` both pass; the
  rewritten descriptor records the named source-cited mutation foundation for
  every deferred row and no working command loses credential-bound dispatch.
- **Refactor/quality:** run the source importer and validator for the target,
  `surface-sync --check`, a focused source-projection test, and an isolated
  built-binary credential-boundary sweep for its implemented commands. The
  shared validator review may still prevent a final batch-wide gate; record the
  exact result without bypassing it.

## Sequential TDD slices

1. **Baseline/red inventory.** Record failing source-import/check results for
   all mutable locks, zero-surface command probes, drifted provider identities,
   Jira’s untyped/read-write accounting, and Stripe’s JSON misclassification.
2. **Source integrity and identity repairs.** Refresh the six mutable sources;
   remove Bitbucket/GitLab stale rows and add Notion’s twelve current rows;
   regenerate descriptors, parameter mappings, and derived surfaces. Red is a
   source-import/check mismatch or set-difference test; green is a verified,
   exact operation identity set.
3. **Terminal surfaces.** Adopt the source-derived, bounded operations needed
   for CircleCI, Sentry, and Vercel commands. Red is a built-binary
   `unknown command`; green dispatch reaches the credential boundary and real
   runtime preflight accepts the command.
4. **Reachability truth.** Convert Jira enabled direct reads to typed operation
   contracts and align its direct-write ledger with executable commands; fix
   Docker Hub metadata/citations and Notion’s comment-action binding. Red is
   an unbound enabled row or contradictory state; green is a typed join and
   credential-boundary probe.
5. **Classification and disabled evidence.** Reclassify Stripe file metadata
   responses as JSON direct reads and regenerate all dispositions so forbidden
   scope reasons disappear and foundation gaps cite present code. Red is the
   independent evidence scan; green is zero forbidden reasons and complete
   citations.
6. **Gate completion.** Run structural checks, focused tests, binary probes,
   GitHub checksums, full `make verify`, then the independent Gate B rerun.
   If it finds gaps, use `plan-phase 4325 --gaps` and
   `execute-phase 4325 --gaps-only` before repeating verification.

## External foundation holds

### Post-#4340 Bitbucket and GitLab disposition slice

- Scope: only `bitbucket` and `gitlab` connector-owned source inventories and
  non-executable mutation dispositions. The shared implementation landed in
  #4340; this slice neither changes shared source-import code nor promotes a
  non-executable mutation to an action or command.
- Red: with the stale inventory masked only for measurement, Bitbucket imports
  297 pinned source operations but source projection reports 29 missing actions
  and validation exposes 77 mutation coverage rows; GitLab imports 1,752 and
  reports 173 missing actions. The temporary zero-row locks are restored
  byte-for-byte after each measurement.
- Green: refresh each existing v2 lock's REST inventory from its already
  pinned descriptor without changing URL, source bytes, or SHA-256; derive one
  exact source-id/method/path disposition for every non-executable mutation;
  rerun source-import and validation. A disposition is a merge-blocking gap,
  never an executable action or `availability: implemented` command.
- Refactor/check: build the real binary and sweep every declared implemented
  Bitbucket and GitLab command in isolated credential-free projects. Each must
  reach `missing --credential`; zero unknown commands are acceptable evidence.

- **Asana:** pending #4326. The provider-owned OpenAPI 3.0 document cannot
  import because `sourceReferenceResolver.referenceTargetWithCount`
  (`cmd/connectorgen/sourceimport.go:5088-5091`) rejects its response `$ref`
  with a `description` sibling. The connector keeps all 249 operations; no
  connector-local omission, downgrade, or shim is permitted.
- **Stripe:** remains pending issue #4323’s cyclic-schema importer work and is
  intentionally sequenced after the independent connector slices.
- **Jira, Notion, and Bitbucket:** are also pending the cyclic-schema importer
  repair in PR #4327. Their live artifacts respectively reach recursive
  `LinkGroup`, `publicApiAsyncTaskStatusResultJsonValue`, and `base_commit`
  schemas at `sourceReferenceResolver.referenceTargetWithCount`
  (`cmd/connectorgen/sourceimport.go:5107-5108`). The held refreshes retain
  their real source sets; no local flattening or fabricated gap is permitted.
- **CircleCI:** pending #4328. Its source-derived webhook actions contain
  `signing-secret`, but `sourceProjectCommand` emits an ordinary reverse-ETL
  shell flag and `checkCLISurfaceEnvOnlyFlags`
  (`cmd/connectorgen/validate.go:896-928`) has no declarative-write secure
  input mode. The held connector surface must not ship before the shared fix.
- **Sentry and Vercel:** pending #4329. A connector with a CLI surface but no
  write actions cannot source-cite unsupported mutations as read-only:
  `validateSourceExecutableCoverage`
  (`cmd/connectorgen/sourceprojection.go:1943-1948`) otherwise requires an
  executable action for every mutation. The solution must be a shared,
  source-cited read-only refusal, not an invented connector gap or `partial`
  command.
- **GitLab and Docker Hub:** are further #4326 consumers. Their OpenAPI
  response references carry a legal `description` sibling, rejected by
  `sourceReferenceResolver.referenceTargetWithCount`
  (`cmd/connectorgen/sourceimport.go:5088-5091`); source refresh and evidence
  regeneration remain held until that shared OAS 3.0 behavior lands.

## Commit and push checkpoints

- Commit this plan and the red baseline separately from production repairs.
- Commit each connector-complete green slice separately, without unrelated
  bundle changes.
- Run full `make verify` serially before every push; never force-push, rebase a
  published branch, or push `main`.

## Safety boundaries

- No provider credential, browser profile, runtime state, reverse-ETL execution,
  destructive provider request, new dependency, test suppression, or quality
  reduction.
- Never edit `internal/connectors/defs/github/rate_limits.json`.
- Keep GitHub lock/descriptor bytes and SHA-256 unchanged; report measured
  values rather than claims.

## Jira direct-write materialization slice (2026-08-25)

This is one connector-only slice under Refs #4325. The independent mapping
manifest identifies three source operations without a current CLI contract:

| Source ID | Provider method/path | Pinned citation | Decision |
| --- | --- | --- | --- |
| `removeGroup` | `DELETE /rest/api/3/group` | `jira-operation-source-lock.json:1489-1495` | Defer: source says neither group selector is individually required, while its `400` response says a group name is required. |
| `resetUserColumns` | `DELETE /rest/api/3/user/columns` | `jira-operation-source-lock.json:4567-4573` | Materialize: two scalar query inputs; no body. |
| `addWatcher` | `POST /rest/api/3/issue/{issueIdOrKey}/watchers` | `jira-operation-source-lock.json:2011-2017` | Defer: required provider body is a JSON string. |

All three retain the source URL, 2,456,011 bytes, SHA-256
`e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf`, and
exact OpenAPI locations in `sources/jira-operation-descriptor.json`. No
operation is relabelled partial. Only `resetUserColumns` becomes an explicit,
closed Jira reverse-ETL action contract with the existing plan → preview →
approval → execute policy and destructive confirmation; no generic HTTP or
write surface is introduced.

`addWatcher` must not be approximated as an object body. Its source declares a
required `application/json` scalar string, while the default JSON executor
builds `map[string]any` (`internal/connectors/engine/write.go:674-692`) and
`WriteAction` has no scalar-JSON body type
(`internal/connectors/engine/bundle.go:507`). Its current
`source-cited-non-executable-mutation-foundation-r1` gap is truthful. The
mapping manifest reserves exact path
`api op-bd7737be09f94f80d0b805cb85032ca567423145501ad839259182b48c939032`
and requires `runtime_deferred_command_projection` with that same path; no
implementation exists in `cmd/` or `internal/`.

`removeGroup` has a separate source-completeness boundary. The pinned artifact
for the same location says response `400` is returned when the group name is
not specified, but OpenAPI marks both `groupname` and `groupId` individually
optional (and both `swapGroup` alternatives optional). A closed record schema
would need an at-least-one and mutual-exclusion rule over query fields; no
declaration-owned action contract has that source-derived conditional-input
projection. Do not guess one required field or send an empty destructive
request. Its reserved manifest path remains unchanged until such a foundation
exists.

- **Red:** add a commandrunner preflight test for manifest-reserved
  `api op-798e4bdcb516fc99a56c6b35b2bc97e67b65830a72dc867eeab1bb261c01b320`;
  it is currently unknown. The no-credential probes for the reserved
  `removeGroup` and `addWatcher` paths return unknown command (exit 2), proving
  absent projections rather than provider failures.
- **Green:** that same manifest-reserved path passes real preflight. Its action
  schema is closed, includes only source-declared query fields, carries no
  body, and a no-credential binary reaches `missing --credential`, never
  unknown command.
- **Hold:** leave `removeGroup` and `addWatcher` deferred until their shared
  conditional-query and scalar-JSON/deferred-command foundations land; do not
  add local shims.
- **Quality:** focused commandrunner test, Jira importer/validator,
  `surface-sync --check`, connector checks, build, probes, help/docs checks,
  and serial `make verify` before a push.

## Asana no-body mutation materialization cohort (2026-08-25)

This connector-owned cohort consumes exactly the 21 Asana source operations
that are already mapped as `declarable` with
`missing_implementation.component: action_binding`. Each has exactly one
required string path parameter, no request body, and, for the 19 DELETEs, an
optional boolean `opt_pretty` query parameter. The source descriptor is the
only input authority:
`internal/connectors/defs/asana/sources/asana-operation-descriptor.json`.

| Source IDs | Method/body shape | Existing canonical CLI paths | Decision |
| --- | --- | --- | --- |
| `approveAccessRequest`, `rejectAccessRequest` | POST, required `access_request_gid`, no body | `access-requests approve-access-request`, `access-requests reject-access-request` | Bind source-derived no-body actions and retain the existing paths. |
| `deleteAllocation`, `deleteAttachment`, `deleteBudget`, `deleteCustomField`, `deleteGoal`, `deleteMembership`, `deleteOooEntry`, `deletePortfolio`, `deleteProjectBrief`, `deleteProjectStatus`, `deleteProjectTemplate`, `deleteRate`, `deleteRole`, `deleteStatus`, `deleteStory`, `deleteTaskTemplate`, `deleteTimeTrackingCategory`, `deleteTimeTrackingEntry`, `deleteWebhook` | DELETE, one required `*_gid`, optional `opt_pretty`, no body | Existing `<resource> delete-<resource>` paths in `cli_surface.json` | Bind source-derived no-body delete actions and retain the existing paths. |

- **Red:** `go run ./cmd/connectorgen source-import asana --check` reports
  descriptor drift (`writes=0 cli=0`), and the planned commands do not pass
  real `commandrunner.Preflight` because they have no named executable action.
- **Green:** each existing canonical command is `implemented`, references its
  matching action, carries only importer-projected path/query flags, and passes
  real preflight. The importer verifies all 249 source operations without
  generating a second command path; the source disposition is removed only for
  an operation after its complete action and implemented command exist.
- **Hard boundaries:** all 21 actions use `body_type: none`; do not invent
  idempotence, missing-status, request schemas, dynamic bodies, provider
  scopes, or a generic write surface. The 24 separate existing Asana
  source-bound request-schema gaps remain intact and no working command is
  downgraded.
- **Quality:** one table-driven commandrunner regression names every source
  ID/action/path triplet; source-import/check, targeted validation,
  `surface-sync --check`, credential-free binary probes, help/inspection,
  vet/build, and `git diff --check` record the real result. Full `make verify`
  remains required before any future push.
- **Green result:** source import refreshed its 249 operation descriptor with
  summaries, then projected exactly 21 write and 21 CLI field updates after
  the action bindings were supplied. The canonical existing paths—not new
  generated paths—pass real preflight. The only post-import Asana validation
  findings are the unchanged 24 source-bound request-schema gaps on other
  implemented operations.
