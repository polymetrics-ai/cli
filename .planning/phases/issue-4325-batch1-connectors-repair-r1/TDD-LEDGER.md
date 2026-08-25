# Issue 4325 — TDD Ledger

| Slice | Red evidence to record before production edit | Green assertion | Refactor/quality gate | Status |
| --- | --- | --- | --- | --- |
| Baseline source drift | `go run ./cmd/connectorgen source-import asana --check` and the CircleCI equivalent each exited 1 because current main lacks `defs/<connector>/sources/` | Re-pinned import and descriptor check pass with provider-derived exact set | `connectorgen validate` and `surface-sync --check` | CircleCI source green / security hold #4328; remaining locks pending |
| CircleCI/Sentry/Vercel surfaces | A built baseline binary returned exit 2 / `unknown command` for all three `operations list` probes | The same command returns `missing --credential` and preflight accepts it | Generated command/help checks | CircleCI pending #4328; Sentry/Vercel pending |
| Jira reachability | Enabled row lacks typed operation or disabled write row contradicts implemented command | Every enabled row has the exact typed operation and direct writes report their actual state | Real preflight sweep | pending |
| Docker Hub/Notion truth | Contradictory status/citation or wrong covered-by action is detectable | Metadata points to existing source/action and current refusal line | Bundle validation | pending |
| Stripe semantics | Four JSON response routes classify as binary | The four routes classify as JSON direct reads and command output policy matches | Semantic regression check | pending |
| Evidence reasons | Report scan finds forbidden scope reason or an uncited foundation gap | Scan finds none; all remaining citations resolve to the stated runtime refusal | Independent Gate B rerun | pending |
| Final gate | Independent report returns NO-GO | Independent report returns GO | Full `make verify` and review | pending |
| Sentry declared-and-deferred mutations | `source-import sentry --check` passes 223 source operations, but `connectorgen validate internal/connectors/defs/sentry` reports 32 exact mutations with no executable action | A strict, source-cited mutation disposition exists for each of those 32; importer and validator accept the descriptor and it records only `source-cited-non-executable-mutation-foundation-r1` gaps | `source-import --check`, targeted validate, `surface-sync --check`, focused source-projection test, and credential-boundary command sweep | green except global surface-sync blocked by unrelated Asana 25-action gap |
| Vercel surface discovery | `pm vercel` returns `unknown command`; an absent `cli_surface.json` means no command can be credential-bound | Report the required declaration-owned artifacts and source-bound operation set before any mutation mapping | Do not create a command surface as part of a disposition-only slice | pending captain report |
| Asana/Jira declared-and-deferred mutations | Targeted validation reports the documented source mutations without complete actions | Exact cited disposition files retain those operations as merge-blocked runtime gaps without changing a working command | Same focused importer/validator and runtime sweep as Sentry | pending |

### 2026-08-24 Sentry declaration-only executable-action conversion

- **Scope and ownership:** Sentry only. This slice changes connector-owned
  `internal/connectors/defs/sentry/**` artifacts plus its focused runtime
  preflight assertion; it introduces no connector-specific runtime or
  generator behavior.
- **Source evidence:** the pinned Sentry descriptor's existing
  `source-cited-non-executable-mutation-foundation-r1` disposition set contains
  exactly 32 operations: 29 `DELETE` and three `POST`. Every one has typed
  required provider path parameters and no request body. Therefore each is
  expressible as a closed, action-owned, no-body direct write; no reverse-ETL,
  delete, request-body, or rendered-reference foundation is needed.
- **Red:** `TestSentryCitedMutationCommandsPassRuntimePreflight` must fail
  against the loaded real bundle because it exposes zero implemented
  reverse-ETL commands for the 32 cited source mutations.
- **Green:** declare all 32 action/command pairs in Sentry definition files,
  with exact source method/path/typed fields, closed record schemas, and
  destructive confirmation. The test must find exactly 32 implemented
  reverse-ETL commands and each must pass real `commandrunner.Preflight`.
  The importer must clear the 32 non-executable disposition gaps without
  removing any source operation, and an isolated no-credential binary must
  reach `error: missing --credential` for every command.
- **Quality:** run focused commandrunner and connectorgen tests, source import
  and validation, connector conformance/surface checks, then the repository
  verification entry points before any push. GSD lifecycle is executed inline
  because this task's single-worker connector contract and the available
  runtime prohibit role spawning; required skills loaded: golang-how-to,
  golang-cli, golang-testing, golang-design-patterns,
  golang-structs-interfaces, golang-error-handling, golang-security, and
  golang-safety.
- **Green result:** `go run ./cmd/connectorgen source-import sentry --check`
  verified 223 provider operations with zero inbound events; targeted
  `connectorgen validate`, commandrunner source-contract tests, and Sentry
  conformance passed. An isolated built binary reached exactly
  `error: missing --credential` for all 32 generated commands (and no command
  was unknown). `go vet ./internal/connectors/commandrunner` and
  `git diff --check` passed. The first full `internal/cli` package run exposed
  stale generated skills for precisely the ten batch-one connectors; after
  `pm skills generate --dir docs/skills`, the deterministic
  `TestSkillsGenerateMatchesTrackedSkills` check passed. A later full
  `internal/cli` run reached the 20-minute package timeout after the unrelated
  `TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess` admission-helper
  failure, but also correctly exposed root-help transcript drift from the
  newly visible CircleCI and Sentry command surfaces. Regenerated exactly the
  nine affected root records, then ran the same selected transcripts green in
  28.283s. The global
  `surface-sync --check` remains blocked by Docker Hub's absent canonical
  descriptor and is not altered by this slice.

### 2026-08-25 CircleCI source-descriptor refresh

- **Red:** `go run ./cmd/connectorgen source-import circleci --check` reports
  descriptor projection drift while `connectorgen validate` is clean. A
  temporary, source-lock-derived descriptor has the same 111 operations and
  retained source artifact identity; its only semantic addition is each
  provider operation summary.
- **Green:** re-run source import against the pinned CircleCI artifact without
  changing the lock, command surface, or actions. The descriptor check and
  targeted validation must pass, and the pre-existing executable commands must
  remain executable at the credential boundary. Do not fabricate additional
  actions for the 83 source-cited merge-blocked operations.
- **Green result:** the refresh added exactly 111 source-provider summaries to
  the descriptor. `source-import circleci --check` verified all 111 operations
  and `connectorgen validate internal/connectors/defs/circleci` found zero
  findings. A fresh built binary, isolated project, and sequential 43-command
  sweep produced exactly `error: missing --credential` for every implemented
  CircleCI command; none was unknown and no field values or credentials were
  supplied.
- **Recorded foundation gap, not a wait:** CircleCI's provider schema describes
  `signing-secret` as a secret in the `createWebhook` and `updateWebhook`
  descriptions (`sources/circleci-operation-descriptor.json:33921-33925` and
  `:35226-35230`) but omits the machine-readable `x-secret` marker. The two
  existing actions redact it, while source projection makes an `env_only` flag
  only when `sourceProjectionDeclaredSecret` sees `x-secret`
  (`cmd/connectorgen/sourceprojection.go:2162-2177`), so a connector-local
  flag edit would be regenerated away. Both commands remain credential-bound
  and working; source-aware secret classification is a shared foundation gap,
  not a reason to downgrade or remove them.

### 2026-08-25 GitLab semantic-POST disposition correction

- **Red:** `go run ./cmd/connectorgen source-import gitlab --check` rejects
  `postApiV4AiThirdPartyAgentsDirectAccess` because the mutation-disposition
  file calls it mutating. The same pinned source's declaration ledger records
  its provider summary, “Get connection details for third party agents,” as a
  lifecycle-restricted semantic POST read.
- **Green:** remove only the false mutation disposition; retain the provider
  operation and its explicit disabled provider-restriction ledger row. Re-run
  source import and validation before declaring any new executable command.
- **Mapped correction:** five—not one—mutation dispositions were stale after
  comparing the full pinned declaration ledger with its semantic-POST read
  set: the AI third-party-agent direct-access, code-suggestions connection
  details, Slack options, and two Conan upload-URL lookup operations. All five
  remain source operations; only their incorrect non-executable-mutation
  labels were removed. Source import refreshed the descriptor with the pinned
  provider summaries and made no action or CLI change.
- **Recorded foundation gap, not a wait:** source import now reaches the one
  genuine missing declaration, `postApiV4CodeSuggestionsConnectionDetails`.
  Its pinned summary is “Retrieve connection details,” has no request body,
  and the declaration ledger classifies it as an implementable semantic POST
  read. `sourceProjectionSummaryDescribesRead`
  (`cmd/connectorgen/sourceprojection.go:1137-1152`) recognizes `get`,
  `list`, `search`, and related verbs but not the equivalent provider verb
  `retrieve`; `sourceProjectionOperationMutates` (`:1124-1132`) therefore
  incorrectly routes it to the action-only path. A connector-local action
  would misrepresent a read, while a local summary rewrite would be regenerated
  from the pinned source. This needs the shared classifier to recognize the
  source verb; leave the source operation and disabled ledger row intact.

### 2026-08-24 Asana red/green split

- **Red:** `go run ./cmd/connectorgen source-import asana --check` reports 25
  source operations without complete executable actions; targeted validation
  reports the same 25 plus 65 separately unresolved source-bound gaps.
- **Classified actionless set:** 21 of the 25 have neither a matching write
  action nor an implemented command. Every one is `direct_write` in the
  existing source-cited declaration-disposition ledger and is eligible for the
  non-executable mutation foundation.
- **Protected executable set:** `deleteProject`, `deleteSection`,
  `deleteTag`, and `deleteTask` each have an implemented reverse-ETL command
  and a matching action. In a fresh credential-free project,
  `pm asana projects|sections|tags|tasks delete` each returned exactly
  `error: missing --credential`. Their actions use a historical `gid` field
  while the pinned provider source requires respectively `project_gid`,
  `section_gid`, `tag_gid`, and `task_gid` (plus optional `opt_pretty`). A
  non-executable disposition would be refused and would be dishonest.
- **Green assertion:** retain those four commands as implemented, update their
  action and command contracts to the exact provider path/query names, and
  defer only the other 21 cited mutations. Source import must pass without
  reducing an executable claim; validation must retain any independently
  source-cited runtime contract gap rather than hiding it.

### 2026-08-24 Asana declared-and-deferred result

- Green mapping: `go run ./cmd/connectorgen source-import asana` updated 69
  declaration-owned write/CLI contracts; its `--check` rerun verified all 249
  provider operations. The 21 actionless `direct_write` rows are each retained
  as a cited `source-cited-non-executable-mutation-foundation-r1` gap. The
  descriptor owns the provider URL, SHA-256
  `cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56`,
  byte count `3066750`, and precise OpenAPI location for every disposition.
- Green executable boundary: after rebuilding `pm`, each of `pm asana projects
  delete`, `sections delete`, `tags delete`, and `tasks delete` returned exit
  1 with exactly `error: missing --credential` from the isolated no-credential
  project. Their paths, schemas, and flags now use the provider names
  `project_gid`, `section_gid`, `tag_gid`, and `task_gid`, plus optional
  `opt_pretty`; none was deferred.
- Focused green: `go test -timeout 20m ./cmd/connectorgen -run
  'TestSourceProjectionSourceCitedNonExecutableMutationDisposition' -count=1`
  passed.
- Honest remaining validation: `connectorgen validate
  internal/connectors/defs/asana` now reports 28 rows, not missing actions:
  24 implemented source operations retain cited
  `cli-request-schema-foundation-r1` gaps for non-scalar `opt_fields` and
  unbounded request bodies, emitted by
  `sourceProjectionOperationParameterGap` / `sourceProjectionSchemaGap` in
  `cmd/connectorgen/sourceimport.go:6826-6900`; four older implemented
  commands retain engine-incompatible backreference regexes. These working
  commands are neither downgraded nor tagged as non-executable.

### 2026-08-24 Asana schema/regex disposition

- Current validation's 24 `cli-request-schema-foundation-r1` rows each carry
  an optional `opt_fields` query gap *and* an unbounded body gap. Only
  `createEnumOptionForCustomField` and `updateCustomField` have optional
  bodies; the Zoom optional-body exception can clear only those two, while
  retaining their citations and counts. The other 22 bodies are required and
  need a complete bounded request contract: `createCustomField`, `createGoal`,
  `createOooEntry`, `createPortfolio`, `createProject`, `createProjectBrief`,
  `createProjectStatusForProject`, `createStatusForObject`, `createTag`,
  `createTask`, `createSubtaskForTask`, `createTeam`, `createProjectForTeam`,
  `createProjectForWorkspace`, `createTagForWorkspace`,
  `updateGoalRelationship`, `updateGoal`, `updatePortfolio`,
  `updateProjectBrief`, `updateTeam`, `updateUser`, and
  `updateUserForWorkspace`. No required body was tagged or made optional.
- The four unsafe-pattern validation rows are two distinct provider patterns,
  each compiled through both the `--data` and `--opt-fields` command flags:
  `duplicateProject` has the 16 project include values and `duplicateTask` the
  10 task include values. Both source patterns use a character class around
  words plus `\\1`; Go's RE2 compiler rejects that backreference at
  `internal/connectors/engine/schema.go:311-314`. They are not a catastrophic
  backtracking risk in this runtime—compilation fails first—but would still be
  functionally unsafe in a backreference engine because `[a|b]` matches one
  character, not one named field. The source-backed correct form is anchored
  `^(?:value1|value2)(?:,(?:value1|value2))*$`, using the documented provider
  value list for each operation. `sourceProjectionSchema` currently copies a
  provider `pattern` verbatim at `cmd/connectorgen/sourceprojection.go:1656-9`,
  so a connector-local rewrite would be regenerated away. This needs a shared
  source-pattern dialect normalization (or provider-source correction), never
  a non-executable tag.
- **Red:** add `TestSourceProjectionNormalizesLegacyDelimitedEnumerationPattern`
  against the exact invalid provider spelling. Before implementation it must
  fail because projection returns the raw `\\1` pattern and the engine schema
  cannot compile it.
- **Green:** a narrowly recognized provider delimited-enumeration form is
  normalized into anchored RE2 alternation, compiles through the engine, accepts
  the documented comma-separated values, and refuses malformed or unrecognized
  source patterns unchanged.
- **Green result:** the focused test first failed with raw
  `([forms|members|task_dates])(,\\1)*`, then passed after shared normalization.
  `source-import asana` regenerated exactly two write contracts and targeted
  validation has zero unsafe-regex findings (24 source-schema gaps remain).

### 2026-08-24 Asana required-body scope

- The 22 required-body findings cover 22 distinct provider routes. Each has a
  single source `data` wrapper, but none supplies a closed bounded schema. Nine
  explicitly retain an open `custom_fields` map (`createProject`,
  `createProjectForTeam`, `createProjectForWorkspace`, `createSubtaskForTask`,
  `createTask`, `updateGoal`, `updatePortfolio`, `updateUser`, and
  `updateUserForWorkspace`); the other 13 retain object schemas whose OAS
  default is also open because they omit `additionalProperties: false`.
- This is provider-source limitation, not a safe connector declaration task:
  closing any of the 22 would omit provider-permitted fields, while retaining a
  generic object would bypass the bounded-record contract. The source has no
  bounded request schema to materialize. Required bodies stay visible failures
  pending a shared, source-cited bounded-dynamic-body capability or a provider
  schema correction; no local schema inference is authorized.

### 2026-08-24 Asana documentation-source audit

- Rechecked every one of the 22 required-body routes against its official
  endpoint page rather than treating the OpenAPI source as exhaustive. All 22
  pages returned HTTP 200 and the provider-supported `.md` renderings are
  pinned, with URLs, byte counts, SHA-256 values, and request/schema pointers,
  in `internal/connectors/defs/asana/sources/asana-required-body-documentation-audit.md`.
- Result: documentation resolves **0/22** to a bounded contract. Every page's
  own OpenAPI definition still leaves its data root (or an `allOf` branch)
  open; nine explicitly retain dynamic `custom_fields` maps and the other 13
  have no `additionalProperties: false` closure. No page says its listed fields
  are exhaustive. This is now a falsifiable two-source provider limitation,
  not an inference from a failed fetch or a connector-local assertion.

## Jira red/green evidence

- Red: `go run ./cmd/connectorgen source-import jira --check` exited 1 because
  no connector-owned source directory exists. The pre-change surface has 590
  commands (584 implemented), 292 implemented direct reads, 286 implemented
  write commands, but only 25 typed operations (22 REST reads and three binary
  downloads). Its 617-row v1 ledger therefore cannot truthfully describe the
  legacy direct-read and direct-write reachability.
- Green assertion: a refreshed source lock and v2 artifact ledger must retain
  the full provider method/path set, bind every implemented direct read to an
  explicit typed operation, and classify direct-write actions according to the
  actual CLI command state. No enabled endpoint may remain operation-less.
- Blocked foundation: after live refresh (2456011 bytes; SHA-256
  `511d0b97390cc47aa0e1367189210a41f32088d9c869e7bb01f43698bdf7e5e8`;
  OpenAPI 3.0.1; 617 operations), `go run ./cmd/connectorgen source-import
  jira` stops on `POST /rest/api/3/issue/bulkfetch`: the recursive
  `#/components/schemas/LinkGroup` reaches
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5107-5108`. Jira is pending the shared
  cyclic-schema importer repair #4327; no local flattening or fabricated gap
  is valid.

### 2026-08-24 Jira declared-and-deferred result

- **Planned red:** add a focused runtime-preflight test for the three
  `universal-avatar` image download commands. It must fail against the loaded
  Jira bundle until each binary operation carries the exact provider success
  status and image media contract.
- **Planned green:** retain the source descriptor's `200` success status and
  `image/png` / `image/svg+xml` media types on each of the three operation
  declarations. This is a bounded provider response contract, not a new
  endpoint or a relaxation of binary-download validation.
- **Red result:** `TestJiraAvatarImageCommandsPassRuntimePreflight` failed with
  `jira.universal_avatar_get_avatar_image_by_type requires non-empty declared
  response content_types`; no request reached a provider.
- **Green result:** after adding only the documented `200`, `image/png`, and
  `image/svg+xml` contracts, the focused runtime-preflight test passed. A
  rebuilt binary in an isolated no-credential project ran each of
  `universal-avatar get-avatar-image-by-type --type project`,
  `…by-id --type project --id 1`, and `…by-owner --type project --entity-id 1`
  to exactly `error: missing --credential` (exit 1), rather than an operation
  metadata refusal.
- **Shared sweep observation:** the real all-bundle preflight sweep now reports
  21 of 5,284 implemented commands failing; every remaining failure is the
  independently-held Docker Hub debt. Jira contributes zero runtime-preflight
  failures. Jira source import intentionally remains at 11 incomplete
  semantic-POST action projections, and targeted validation reports 318
  historic projection findings but zero avatar response-contract findings.

- Red: current `go run ./cmd/connectorgen source-import jira --check` reported
  14 incomplete source-action contracts; current targeted validation reported
  16 source operations without an executable action, 86 unresolved source
  gaps, and further source request-field findings.
- Classified deferred writes: `removeGroup`, `resetUserColumns`,
  `DynamicModulesResource.removeModules_delete`, and `addWatcher` have no
  matching write or implemented CLI command. Each is `direct_write` in the
  existing Jira declaration disposition ledger. The new four-row mutation
  disposition file lets source import retain exact provider citations from the
  current descriptor: URL
  `https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json`,
  SHA-256 `e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf`,
  byte count `2456011`, and its individual OpenAPI location; all are
  merge-blocking `source-cited-non-executable-mutation-foundation-r1` gaps.
- Protected direct reads: 11 remaining source-action findings are semantically
  `direct_read` POST endpoints, each with an implemented command. A rebuilt
  no-credential binary reached exactly `error: missing --credential` (exit 1)
  for `changelog get-bulk-changelogs`, `comment get-comments-by-ids`, `issue
  get-change-logs-by-ids`, `jql get-auto-complete-post`, `jql
  get-precomputations-by-id`, `jql match-issues`, `jql migrate-queries`, `jql
  sanitise-jql-queries`, `search count-issues`, `workflow
  list-workflow-history`, and `worklog get-worklogs-for-ids`. The related
  `app get-custom-fields-configurations` direct-read command reached the same
  boundary and is already excluded from the missing-action set.
- Re-derived credential split: Jira declares 584 implemented commands (292
  direct reads and 286 write commands among them). The real
  `TestEveryImplementedCommandPassesRuntimePreflight` sweep reached 581
  credential boundaries; its only three Jira failures are the universal-avatar
  image-by-type, image-by-id, and image-by-owner binary-download operations,
  each missing declared response `content_types`. The 26/5,284 global failure
  result also includes the independently recorded Asana and Docker Hub debt;
  no test was weakened or bypassed.
- Foundation stop: `sourceProjectionOperationMutates` at
  `cmd/connectorgen/sourceprojection.go:1116-1121` defines all REST POST
  operations as mutations. It therefore makes those 11 working direct reads
  enter the action-only projection path; `sourceProjectionReadOnlyDeclaration`
  at `:1084-1096` cannot represent them because its only model is a blocked
  read-only endpoint. Neither a non-executable mutation disposition nor a
  connector-local action is truthful. After the four cited write dispositions,
  `source-import jira` honestly remains at 11 source-action findings. This
  requires a shared source-projection semantic-POST direct-read contract.

### 2026-08-24 Jira semantic-POST classification

- **Red:** add a classifier test with provider-owned operation summaries: a
  POST whose source summary says `Search widgets` must be a read, while a POST
  whose source summary says `Create widget` must remain a mutation. Add an
  importer assertion that the operation summary survives in the immutable
  descriptor. Before the classifier change, a POST is always classified as a
  mutation (and the descriptor carries no source semantic summary).
- **Green assertion:** retain the OpenAPI operation summary in the source
  descriptor, and use that provider-described semantic only to distinguish an
  otherwise-mutating REST POST. No connector-local override or operation list
  is permitted. The 11 Jira query/search/conversion POST commands must remain
  executable direct reads and reach the missing-credential boundary.
- **Green result:** `TestSourceImportRetainsOperationSummary` and
  `TestSourceProjectionClassifiesSemanticPOSTOperations` pass. The descriptor
  now retains the OpenAPI `summary`; a `Search widgets` POST is a read, while
  a `Create widget` POST remains a mutation. `source-import jira --check`
  verifies all 617 provider operations, and Jira no longer reports a
  source operation without an executable action. Targeted validation retains
  only 23 independently source-cited request-schema gaps on genuine write
  operations.
- **Credential boundary:** after rebuilding `pm` and initializing a fresh
  credential-free project, all 11 affected commands—`changelog
  get-bulk-changelogs`, `comment get-comments-by-ids`, `issue
  get-change-logs-by-ids`, all five `jql` commands, `search
  count-issues`, `workflow list-workflow-history`, and `worklog
  get-worklogs-for-ids`—each exited 1 with exactly `error: missing
  --credential`. None was downgraded, deferred, or replaced by a write action.

### 2026-08-24 captain-directed batch-one operation split

- **Independent foundation measurement:** this branch contains 2,292
  `reverse_etl` commands across 29 connectors, 1,999 `direct_read` commands
  across 22, 658 `etl` commands across 36, 27 `binary_download` commands
  across six, and 285 `direct_write` commands across two. These are this
  branch's figures, not an attempt to reconcile concurrently changing lanes.
  A real `TestEveryImplementedCommandPassesRuntimePreflight` sweep checked
  5,262 implemented commands and found 21 failures, all Docker Hub; no other
  batch-one connector failed runtime preflight.
- **Measured source-operation denominator:** the ten current source locks
  retain 4,341 provider REST operations. The three-way split is **767 already
  runnable**, **1,666 declarable now**, and **1,908 genuinely blocked**. The
  runnable number starts from exact source-to-command crosswalk bindings and
  excludes Docker Hub's two crosswalk-bound commands that fail preflight. A
  declarable operation has no source gap except
  `closed-source-operation-execution-foundation-r1` and, where present, the
  now-known-to-be-declaration-only
  `source-cited-non-executable-mutation-foundation-r1`; it therefore needs
  only a connector-owned action/CLI declaration. Every other source gap,
  including an absent descriptor after an importer refusal, is counted blocked
  until an existing declaration contract can represent it.
- **Per connector (runnable / declarable / blocked):** Docker Hub 4 / 0 / 50;
  Notion 42 / 0 / 7; Stripe 8 / 0 / 581; Bitbucket 50 / 136 / 111; GitLab 0 /
  917 / 835; CircleCI 16 / 75 / 20; Sentry 3 / 144 / 76; Vercel 0 / 261 /
  139; Asana 82 / 130 / 37; Jira 562 / 3 / 52. The Docker Hub, Notion, and
  Stripe blocked counts are their respectively 54, 49, and 589 locked
  operations less the still-working currently declared commands: each source
  import is refused before a descriptor can be produced (dangling
  `#/components/responses/team_repo`, descriptor-byte limit, and schema-depth
  limit respectively).
- **Sentry correction:** all 32 source-cited non-executable mutation
  dispositions have provider-documented path fields (and no request body).
  All 32 are declaration work, not a missing reverse-ETL or delete foundation.

Every red/green command and its result is appended when it executes. No test is
weakened or skipped to advance a row.

## 2026-08-25 — source declaration admission design (captain override)

### Scope and invariant

Every operation retained by a connector's provider source lock must have a
source-cited declaration regardless of whether the operation can execute now.
The declaration must retain its exact method/path/protocol, provider URL,
SHA-256, byte count, document location, semantic class, and destructive or
secret metadata. A non-runnable declaration is `deferred` with a named,
source-cited foundation gap; it is not an `implemented` command and it is not
removed from the source denominator.

This is intentionally separate from runtime preflight. `implemented` remains
an executable claim and must retain real `commandrunner.Preflight` and
credential-free built-binary evidence. A deferred source row must never be
promoted merely to satisfy the declaration certificate.

### Red cases to preserve

1. A locked provider operation with no declaration currently fails source
   projection/validation because it has no complete executable action or
   reachable runtime operation (`sourceprojection.go:187-290` and
   `:2430-2524`). That rejects coverage even when the JSON can honestly carry
   a deferred foundation reference.
2. `batch gate` excludes a complete source bundle with zero implemented
   commands (`batch.go:939-956`), so it cannot be the source-admission
   certificate for a deliberately deferred-only connector.
3. An operation-evidence row with no enabled command receives
   `runtime_reachability` and `cli_command` gaps
   (`operationevidence.go:743-752`); its fixed-100 selection requires both a
   declared and enabled classification (`:1200`). These are valuable runtime
   facts, but must not erase the source declaration.

### Green contract and smallest safe change

Add a declaration-completeness certificate that consumes the retained source
identity set (or its source descriptor where importable), crosswalk,
declaration-disposition ledger, and `api_surface.json`. It passes only when
every source identity is represented exactly once after documented base-path
normalization, with either:

- a closed, runtime-eligible operation/action/command contract; or
- an explicit non-runnable declaration containing the exact source trace,
  classification, risk/destructive metadata, and source-cited refusing
  foundation.

It must fail on an omitted/duplicate source identity, stale unnormalised
endpoint, lost method/path, missing source citation, missing destructive
metadata, a fabricated foundation citation, or a deferred row marked
`implemented`. Existing runtime enforcement remains unchanged:
`TestEveryImplementedCommandPassesRuntimePreflight`, built-binary
`missing --credential` sweeps, and the engine's closed binding/body checks are
still the only proof of executability.

`surface-sync` is not changed to invent commands—it synchronizes existing
operation-owned command metadata. `surface-reconcile` remains runtime-only:
it may mark an endpoint blocked when preflight fails, but cannot define source
coverage. `certification-sweep` may retain its non-implemented N/A accounting,
and operation evidence may retain runtime gaps; neither result is allowed to
be treated as an absent source declaration.

### Test matrix before implementation

| Case | Admission certificate | Runtime/preflight assertion |
| --- | --- | --- |
| Complete runnable REST read | passes exact source-to-declaration join | implemented command passes real preflight and built binary stops at missing credential |
| Deferred mutation/delete | passes with exact source trace, destructive metadata, and cited foundation | no implemented command/action is accepted; executing a deferred command returns its declared block |
| Deferred binary operation | passes with content/response metadata plus foundation citation | binary preflight still refuses unsafe or incomplete metadata |
| Provider dialect/importer refusal | emits/retains an operation-level source gap rather than dropping the descriptor | no runnable command is generated from the incomplete shape |
| Missing, duplicate, stale, or path-normalization-mismatched row | fails with source ID and file:line | no runtime assertion is substituted for coverage |
| Falsely implemented deferred row | fails admission and existing runtime sweep | `TestEveryImplementedCommandPassesRuntimePreflight` remains unchanged |
| Zero-runnable connector (Vercel-shaped) | admits complete deferred provider coverage | batch/runtime certification reports deferred, not source omission |

No production certification or runtime code is changed in this slice. The
captain-directed definition-only conversion continues only after this design
record is reviewed against the report below.

## CircleCI red/green evidence

- Red: the baseline `pm circleci operations list` exited 2 with
  `error: unknown command "circleci"`.
- Green source: the 2026-08-23 provider retrieval measured 621321 bytes,
  SHA-256 `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`,
  and 111 exact method/path identities. `go run ./cmd/connectorgen source-import
  circleci --check` passed.
- Green ledger: the temporary batch materializer accepted the same live artifact
  with 111 artifact operations and no dropped candidate. The retained v2
  ledger has 40 executable bindings and 71 source-cited blocked/disallowed
  rows. `go run ./cmd/connectorgen validate internal/connectors/defs/circleci`
  and `go run ./cmd/connectorgen surface-sync --check` both passed.
- Green runtime: from a fresh initialized project with no credential,
  `pm circleci contexts list` and `pm circleci delete context apply
  --context-id fixture-context` each returned exactly `error: missing
  --credential` (exit 1), not `unknown command`. No provider request was made.

## Foundation holds

- Asana is pending #4326, not partial. The real source-import failure is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`; its OpenAPI 3.0 response `$ref`
  has a `description` sibling. The shared importer must retain the source-bound
  contract before this slice can continue.
- Stripe is pending #4323’s source-import cyclic-schema repair and stays last.
- CircleCI webhook writes are pending #4328. The locked source declares
  `signing-secret`, but `sourceProjectCommand` at
  `cmd/connectorgen/sourceprojection.go:1550-1592` generates an ordinary
  reverse-ETL CLI flag. `checkCLISurfaceEnvOnlyFlags` at
  `cmd/connectorgen/validate.go:896-928` permits secure env-only input only
  for operation-backed GraphQL contracts. The shared write-field secret-input
  capability or a source-bound gap is required; no connector-local shim or
  partial label is valid.
- Sentry is pending a shared read-only source-projection capability. With the
  generated four-command CLI surface present, `connectorgen validate
  internal/connectors/defs/sentry` reports 34 `source operation has no
  executable action` findings. The enforcement is
  `validateSourceExecutableCoverage` at
  `cmd/connectorgen/sourceprojection.go:1943-1948`: a source mutation with no
  action is valid only when the importer itself generated a blocking runtime
  gap. These are unsupported reverse-ETL operations in an intentionally
  read-only connector, not a source grammar failure, so an invented connector
  gap, action, or `partial` command would be false evidence. A shared,
  source-cited read-only refusal model is required.
- Vercel is pending the same #4329 foundation. The checked-in source lock has
  400 operations (163 reads and 237 mutations). Its 18 writes match only four
  source method/path identities, and its CLI surface is absent. Adding a
  surface before #4329 would produce the identical dishonest all-or-nothing
  mutation demand, so the connector remains untouched.

## Notion red/green evidence

- Red source: after importing the prior source envelope, live retrieval found
  61 OpenAPI operations, not the 49 locked rows (1304814 bytes, SHA-256
  `dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258`).
- Blocked foundation: `go run ./cmd/connectorgen source-import notion` stops
  at `GET /v1/async_tasks/{task_id}` on the recursive
  `#/components/schemas/publicApiAsyncTaskStatusResultJsonValue`, refused in
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5107-5108`. The shared #4327 cycle-gap
  importer repair is needed; no connector-local flattening or invented gap is
  honest.

## Bitbucket red/green evidence

- Red source: live retrieval found a 297-operation OpenAPI document, replacing
  the former 331-row lock (1359673 bytes, SHA-256
  `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3`).
- Blocked foundation: `go run ./cmd/connectorgen source-import bitbucket`
  stops at `GET /repositories` on the recursive
  `#/components/schemas/base_commit`, refused at
  `cmd/connectorgen/sourceimport.go:5107-5108`. The shared #4327 cycle-gap
  importer repair is required before the stale source rows can be truthfully
  removed and the source projections regenerated.

## GitLab red/green evidence

- Red source: live retrieval found 1752 operations after the documented
  `/api/v4` base normalization, instead of the former 1755 lock rows
  (3576860 bytes, SHA-256
  `6b6ad591ff1b54ab429d0502812a2b2955501f1f6bebdae1888ba0bea086cf82`).
- Blocked foundation: `go run ./cmd/connectorgen source-import gitlab` stops
  at `POST /api/v4/glql`: response 200 has a `$ref` with a `description`
  sibling. The importer refusal is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`. GitLab needs #4326 before
  stale source rows and its 981 foundation-gap citations can be regenerated.

## Docker Hub red/green evidence

- Green source retrieval: Docker Hub remains exact at 54 operations, 148322
  bytes, and SHA-256
  `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`.
- Blocked foundation: `go run ./cmd/connectorgen source-import dockerhub`
  stops at `POST /v2/auth/token`, response 401, because its `$ref` has a
  `description` sibling. The failure is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`. Docker Hub needs #4326 before
  the source projection can truthfully repair stale citations and contradictory
  enabled/blocked ledger metadata.

## Post-#4327 red/green revalidation

- Green foundation observation: after merging `origin/main` at `02a2201ed`
  (including #4327 at `e338cd301`), none of the seven former cycle or
  descriptive `$ref` sibling errors recurred.
- Re-run: `source-import` passed CircleCI and Sentry only (2/10). The full
  `connectorgen validate` sweep passed CircleCI only (1/10).
- New red: Asana (25) and Jira (16) reach source projection but have incomplete
  executable action contracts (`cmd/connectorgen/sourceprojection.go:211`).
  Bitbucket and Notion reach `schema depth limit exceeded`
  (`cmd/connectorgen/sourceimport.go:4271`); Stripe reaches `reference depth
  limit exceeded` (`cmd/connectorgen/sourceimport.go:5170`). Docker Hub has an
  unresolved source reference (`cmd/connectorgen/sourceimport.go:5496`) and
  existing SCIM `example` keywords rejected by the engine dialect
  (`internal/connectors/engine/schema.go:168`). GitLab's provider path has a
  placeholder with no required parameter (`cmd/connectorgen/sourceimport.go:6048`).
  Vercel's live 400-operation source rejects OAS 3.0 `patternProperties`
  (`cmd/connectorgen/sourceimport.go:4314`). Sentry remains blocked by its 34
  read-only mutation coverage findings (`cmd/connectorgen/sourceprojection.go:1943-1948`).
- No failure is a rendered-reference contract request: every observed error is
  in the OpenAPI importer/projection/engine path after source retrieval.

## Post-#4340 Bitbucket and GitLab source-cited mutation dispositions

- Red: After #4340 merged, temporarily masking each stale REST inventory while
  preserving its source artifact identity lets the real importer reach the
  projection gate. Bitbucket has 29 source-import missing actions and 77 total
  source executable-coverage findings; GitLab has 173 for both. The locks were
  restored byte-for-byte after the measurements.
- Green assertion: connector-owned `*-mutation-dispositions.json` files cite
  the exact source ID, method, and path for every retained non-executable
  mutation. Source-import accepts the refreshed 297/1,752 inventories and
  validator coverage accepts only merge-blocking, non-implemented gaps.
- Behavioral assertion: every remaining command declared `implemented` in the
  Bitbucket or GitLab bundle dispatches in an isolated built binary to exactly
  `error: missing --credential`; zero commands may be unknown.
- Green source: Bitbucket refresh retained its pinned 1,359,673-byte / SHA-256
  `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3`
  artifact, replaced 331 stale lock rows with 297 exact source identities,
  added 29 exact non-executable mutation dispositions, and source projection
  reconciled 49 write/CLI contracts. GitLab retained 3,576,860 bytes /
  `6b6ad591ff1b54ab429d0502812a2b2955501f1f6bebdae1888ba0bea086cf82`,
  refreshed 1,755 rows to 1,752, and accepted 173 exact dispositions. Each
  connector validates with zero findings.
- New red / foundation stop: Bitbucket source projection creates 50
  `implemented` command paths, 28 containing literal `{parameter}` segments.
  An isolated built binary rejects those at command parsing with `error:
  command path segment 4 contains invalid character '{'`, before credential
  preflight. `sourceProjectionGeneratedCommandPath`
  (`cmd/connectorgen/sourceprojection.go:1871-1874`) must derive runtime-valid
  segments rather than emitting raw source IDs; no existing foundation owns the
  shared generator/runtime contract, so connector-local path edits would be
  regenerated and are forbidden.

## Sentry red/green evidence

- Red: the baseline `pm sentry operations list` exited 2 with
  `error: unknown command "sentry"`.
- Red source: `go run ./cmd/connectorgen source-import sentry --check` first
  exited 1 because the imported lock still carried the retired
  `rest.operation_counts` field.
- Green source: after normalizing the lock envelope, the live provider
  document measured 3868570 bytes and SHA-256
  `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`.
  It has 223 operations (120 GET, 35 DELETE, 2 PATCH, 34 POST, and 32 PUT),
  and `go run ./cmd/connectorgen source-import sentry --check` passed.
- Next green assertion: batch materialization must preserve all 223 cited
  endpoint rows while exposing only the four existing read streams. The absent
  `writes.json` is intentional: no reverse-ETL action exists or is claimed.
- Red hook contract: a new focused test must require `projects` to resolve to
  the provider's organization-scoped path and reject a missing organization;
  it must fail against the former legacy-wide `/api/0/projects/` implementation.

### 2026-08-24 declared-and-deferred green evidence

- Added 32 exact source-id/method/path rows in
  `sources/sentry-mutation-dispositions.json`. The only source class for all
  32 is `direct_write`, already recorded with the same provider source URL,
  SHA-256 `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`,
  byte count `3868570`, and OpenAPI location in
  `sentry-declaration-disposition.json`.
- Green: `go run ./cmd/connectorgen source-import sentry`, then
  `go run ./cmd/connectorgen source-import sentry --check`, and
  `go run ./cmd/connectorgen validate internal/connectors/defs/sentry` all
  passed. The regenerated descriptor contains 32 cited
  `source-cited-non-executable-mutation-foundation-r1` merge-blocking gaps.
- Focused proof: `go test -timeout 20m ./cmd/connectorgen -run
  'TestSourceProjectionSourceCitedNonExecutableMutationDisposition' -count=1`
  passed. A fresh no-credential project built from the current binary ran
  `pm sentry events|issues|projects|releases list`; each stopped exactly at
  `error: missing --credential`, with zero `unknown command` results.
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
  remains blocked before Sentry because Asana has 25 source operations without
  complete executable actions. A target-directory invocation is not a valid
  replacement because it lacks its sibling connector fixtures. No gate was
  suppressed.
