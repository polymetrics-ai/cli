# TDD Ledger — #4352 source-bound read execution foundation

## Repair r4 frozen red/green evidence

| Finding | Red before implementation | Required complete green result |
| --- | --- | --- |
| AUDIT-001 | Assert the current source projection emits raw `offset`/colliding `limit` for a source-bound read. | Shared paging classifier excludes provider navigation controls and regenerated surface/help/direct APIs expose only closed derived paging. |
| AUDIT-002 | Supply cross-origin source-bound config through CLI/App and observe credential/auth construction before error. | Same input returns origin error with zero vault/credential reads, auth cohort/protected state, and requester calls. |
| AUDIT-003 | Alter one direct stream source-bound field while retaining its binding and observe direct `Read` reach the later path. | `Read` and `ReadWithOutcome` reject source ID/method/path/records/pagination/origin drift before auth/I/O. |
| AUDIT-004 | Census source-complete Asana mutations and show the 19 DELETE + two no-body POST commands are blocked. | Existing reverse-ETL/delete route admits all 21 to `missing --credential`, while true named gaps remain declared blocked. |
| AUDIT-005 | Run both isolated fixed-100 checks and retain their missing-Asana-input failure. | Copy exact required inputs; both checks and full `./cmd/connectorgen` are green. |
| AUDIT-006 | Assert source-derived Asana facts against stale generated docs/manual/help/website artifacts. | Regeneration and semantic tests show actual counts, closed paging, and only real named foundations. |

Record exact commands, intended red failure, and full green result as each group completes; a passing command alone is not proof.

### AUDIT-005 completed

- **Red:** `go test -timeout 20m ./cmd/connectorgen -run 'TestOperationEvidence(Fixed100RejectsEveryRegression|CheckRunsFixed100Gate)$' -count=1` failed at the immutable review head with `asana.rest.getCustomFieldsForWorkspace is absent from operation evidence`.
- **Intermediate red:** after copying only the Asana bundle, the same test failed `execution evidence regressed`, proving the isolated workspace also needs Asana's website evidence rather than an implicit repository-tree dependency.
- **Green:** the isolated workspace now copies the referenced Asana bundle and carries the Asana + GitHub website rows. The exact command passed in `6.906s`; mutation controls inside each fixed-100 test remain negative proof that a changed expected digest still fails.

### r4 six-finding red/green completion

- **AUDIT-001 Red:** the source projection fixture emitted `query.limit`; the
  retained source-bound surface also exposed provider paging controls. **Green:**
  `TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL`,
  `TestSourceBoundReadsRejectRawProviderPagingControls`, and
  `TestSourceBoundReadHelpUsesClosedPagingFlags` pass; source import and help
  show no source-owned `offset`/`limit` flag and derived navigation remains.
- **AUDIT-002 Red:** invalid source origin was admitted only after later app
  work. **Green:** `TestSourceBoundOriginRejectsBeforeAppOrCredential` uses an
  uninitialised project and observes the source-origin error, never `missing
  project` or `missing --credential`; engine/commandrunner preflight tests
  pass with public configuration only.
- **AUDIT-003 Red:** `Read` had no source-bound stream route proof. **Green:**
  `TestReadRejectsSourceBoundStreamDriftBeforeOriginOrAuthentication` rejects
  method, path, records, and pagination substitutions before origin/auth.
- **AUDIT-004 Red:** source import/validation exposed 21 eligible Asana
  mutations without an action, and the first promotion attempt failed
  validation because command-owned `approval.confirm` flags were invalid.
  **Green:** the action materializer removes that duplicate mapping; all 94
  actions execute through the real reverse-ETL test path and the source test
  proves exactly 19 DELETE + 2 POST endpoint bindings.
- **AUDIT-005 Red/Green:** recorded above; the final full
  `go test -timeout 20m ./cmd/connectorgen` additionally completed with pass.
- **AUDIT-006 Red:** `pm docs validate --connectors-dir docs/connectors`
  reported stale catalog output. **Green:** `pm docs generate`, website data
  generation, documentation validation, and the 34 website script tests pass;
  generated help shows the promoted delete as `implemented`.

### Fresh-audit follow-up red/green ledger

- **F3 / AUDIT-001 Red:** the independent audit exercised a source-promoted
  Asana next-URL list and found that the generated operation omitted its
  provider-declared `limit`/`offset` pagination contract. Page one named no
  `limit`, and a returned continuation was refused as an undeclared query
  parameter. **Green required:** an engine regression proves page one sends
  the derived bounded `limit`, page two is reached only through
  `--page-cursor`, and no raw paging flag is exposed.
- **F4 / AUDIT-006 Red:** eight source-complete implemented reads rendered a
  contradictory historical `Blocked until …` note. **Green required:**
  projection removes that legacy blocker during source-complete promotion and
  generated manual/help/website output contains no such note on an implemented
  source-bound command.
- **F5 / AUDIT-006 Red:** `go test -timeout 20m -count=1 ./internal/cli -run
  '^TestSkillsGenerateMatchesTrackedSkills$'` fails on tracked
  `docs/skills/pm-asana/SKILL.md`, which still exposes raw `--limit` and
  `--offset`. **Green required:** regenerate through `pm skills generate`,
  then the tracked-skill test and docs/website checks pass.

### Follow-up completed red/green evidence

- **F3 / AUDIT-001 Red:** the independent review found that an otherwise
  source-complete next-URL list request sent no `limit`, then rejected Asana's
  returned `limit`/`offset` continuation. **Green:**
  `TestOperationDirectReadNextURLUsesClosedLimitOffsetContinuation` passes:
  the first request uses `limit=100`, the cursor reaches page two, and no raw
  pagination CLI flag is introduced.
- **F4 / AUDIT-006 Red:** eight promoted reads retained `Blocked until …` in
  their generated command notes. **Green:** source-complete promotion clears
  that historical note; the final JSON scan over all source-bound direct reads
  returns an empty result.
- **F5 / AUDIT-006 Red:** the full CLI suite failed because the tracked Asana
  skill exposed old raw paging flags. **Green:** `pm skills generate`, docs
  generation, and website data generation refreshed tracked output; the final
  full CLI package passed in `420.427s`.
- **Boundary/lint Red:** generic source projection still selected behavior by
  connector name and retained dead helpers. **Green:** source-gap annotation
  is connector-neutral, `make connector-boundary` reports no findings, and
  `make lint` reports 0 issues.

### Fresh-audit F6 / AUDIT-002 red/green

- **Red:** independent `codex review --base b33983927d863032dac8220949990506e812937d`
  at repair SHA `07251df15c904cad0f91a43724810dffa133b81d` found that a
  persisted credential `base_url` bypassed the CLI source-origin preflight and
  reached `ResolveConnectorCredential` before rejection.
- **Green required:** a persisted credential configuration is read through a
  credential-only, no-vault/no-App public snapshot; its configuration plus the
  command overlay reaches `PreflightSourceBoundOrigin` before credential
  resolution. The command must reject the persisted invalid origin rather than
  report a credential, vault, authentication, or provider result.
- **Green:** `go test -timeout 20m -count=1 ./internal/cli -run
  '^(TestSourceBoundOriginRejectsBeforeAppOrCredential|TestSourceBoundOriginRejectsPersistedCredentialConfigBeforeVault)$'`
  passed. The new persisted-configuration case first failed on the unmodified
  behavior with `read encrypted credential`, then passed with the declared
  source-origin rejection.

### Current-main integration #4351 red/green

- **Red:** before integration, `git merge-base HEAD origin/main` was
  `b33983927d863032dac8220949990506e812937d`, not the Captain-authorized
  `origin/main` `1324c52bab0b224ed8958858af7676b8b8e191b4`; therefore this
  branch did not contain the merged declaration-admission foundation.
- **Green required:** merge current `origin/main` without force-push or PR
  merge; resolve only actual conflicts; then rerun the actual isolated Asana
  credential-boundary census and the source-bound read/delete/ETL validation
  gates. Record the exact merged SHA and results.

## Lifecycle

- GSD source resolution: `scripts/gsd doctor`, all five required `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check` passed before planning.
- Execution mode: inline/manual. Compatible isolated GSD runtime agents are unavailable and the repository's canonical delivery contract forbids spawning role agents.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`.

| Slice | Red evidence | Green evidence | Refactor / status |
| --- | --- | --- | --- |
| Source projection | `go test ./cmd/connectorgen -run '^TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL$' -count=1` failed: `get_access_requests` had no source binding | `go test ./cmd/connectorgen -run '^(TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL|TestSourceProjectionLeavesIncompleteReadAsNamedFoundation|TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical)$' -count=1` passed | Non-mutating GET descriptors now bind only an existing exact declaration; a nonempty author-owned foundation note is preserved. Green |
| Closed runtime binding | New route/identity test initially had no `SourceOperation` model or engine preflight. | `go test ./internal/connectors/engine -run '^TestPreflightSourceBound' -count=1` and `go test ./internal/connectors/commandrunner -run '^TestRunSourceBoundOperationDirectReadRejectsBeforeDispatch$' -count=1` passed | Binding is `source_operation` ID + GET method + relative path, checked before dispatch. Green |
| Direct versus stream semantics | The red projection fixture had no direct-read promotion or source-backed stream classification. | Source projection green fixture proves bounded `/access_requests`, path-bound `/agents/{agent_gid}`, and paginated `workspaces`; `go test ./internal/connectors/defs/asana -run '^TestSourceBoundReadControlsReachEnginePreflight$' -count=1` passed | Only a named stream with record schema and pagination receives ETL treatment. Green |
| Missing foundation | No stable source-bound reason existed; an incomplete typed path could fall through to generic operation blocking. | `go test ./cmd/connectorgen -run '^TestSourceProjectionImportsRequiredSourceBoundReadParameters$' -count=1` and `go test ./internal/connectors/commandrunner -run '^TestRunSourceBoundReadMissingFoundationRefusesBeforeDispatch$' -count=1` passed | A required scalar contract is imported only from locked source material; unsupported contracts retain stable `missing_foundation` before dispatch. Green |
| Credential boundary | Valid new Asana read was not callable. | A fresh `go build -o <temp>/pm ./cmd/pm`, `pm init --root <temp> --json`, then `pm asana access-requests get-access-requests --root <temp> --json` returned `missing --credential` (exit 1), with no credential or provider request. | Valid direct read reaches normal credential admission. Green |
| Safety regression | New source-bound stream preflight initially incorrectly applied to all legacy ETL commands; the full runner suite exposed the regression. | Restored legacy ETL routing and ran `go test -timeout 20m ./internal/connectors/commandrunner -count=1`, `go test -timeout 20m ./internal/connectors/engine -count=1`, and `go test -timeout 20m ./internal/connectors/defs/asana -count=1`, all passed. | Direct-write, reverse-ETL, binary, delete, and legacy ETL controls retain their existing paths. Green |
| Audit repair write-surface guard | `go test -timeout 20m -count=1 ./internal/connectors/defs/asana -run '^TestReverseETLWriteActionsExecute$'` failed after a retained-source import rewrote existing write actions to a generic required `--data` flag and rejected legacy records before dispatch. | `go test -timeout 20m -count=1 ./cmd/connectorgen -run '^TestRetainedAsanaSourceImportRejectsReadProjectionDrift$'` passes while asserting byte-identical `writes.json` plus every reverse-ETL/delete CLI command; `go test -timeout 20m -count=1 ./internal/connectors/defs/asana -run '^(TestReverseETLWriteActionsExecute|TestDestructiveOperationsStayBlocked)$'` passes. | The new `--read-projection-only` source-import lane seals only pre-existing implemented direct reads/ETL streams. It neither promotes planned GET commands nor changes write/delete semantics. Green. |
| Audit F1 — generated help | `POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES='root_bare_manual,root_long_help,root_short_help,root_help_command,root_man_command,root_json_help,root_late_json_help,root_equals_form,root_space_form' go test -timeout 20m -count=1 ./internal/cli -run '^TestGoldenTranscripts$'` failed on each Asana root-help summary; `go test -timeout 20m -count=1 ./internal/cli -run '^TestSkillsGenerateMatchesTrackedSkills$'` failed with generated skill drift; `go run ./cmd/pm docs validate --connectors-dir docs/connectors` reported stale Asana manual. | Regenerated and reviewed only `docs/skills/pm-asana/SKILL.md`, `docs/connectors/asana/{MANUAL,SKILL}.md`, and the nine root transcript records. The same selected golden test, `TestSkillsGenerateMatchesTrackedSkills`, and `pm docs validate` passed. | Generated artifacts now state bounded source-bound Asana GET access without inventing an HTTP escape hatch. Green |
| Audit F2 — pinned source proof | The retained-source test initially failed: its mutated workspace pagination returned source-import `--check` exit 0, proving a source-bound ETL command could retain `implemented` after its pagination semantics diverged. Its first repair then marked a newly eligible GET as deferred before projection, masking source identity drift; the next repair over-projected 103 planned GETs. | `go test -timeout 20m -count=1 ./cmd/connectorgen -run '^(TestRetainedAsanaSourceImportRejectsReadProjectionDrift|TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL|TestSourceProjectionImportsRequiredSourceBoundReadParameters)$'` passes. It imports only the retained bytes, rejects invented source identity, method, route, typed query input, and workspace pagination, and proves a planned direct read remains unbound. `go run ./cmd/connectorgen source-import asana --read-projection-only --check` passes. | The retained v3 lock and content-addressed artifact are the sole test input. The two-phase read-only projection first seals a pre-existing executable declaration, then records unresolved source reads; planned Batch-1 GETs remain declaration-pending. Green. |
| Captain mutation reconciliation | `go test -timeout 20m -count=1 ./cmd/connectorgen -run '^TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation$'` failed: retained Asana declared zero source mutation dispositions for 25 source routes with no complete exact action. The pre-existing `TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsImplementedIncompleteActionClaim` passed, proving that the old non-executable disposition could not truthfully cover the 65 implemented reverse-ETL operations. | `go test -timeout 20m -count=1 ./cmd/connectorgen -run '^(TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation|TestSourceProjectionSourceCitedPartialMutationCoveragePreservesImplementedIncompleteAction|TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsImplementedIncompleteActionClaim)$'` passes. `go run ./cmd/connectorgen source-import asana --read-projection-only --check`, `go run ./cmd/connectorgen validate internal/connectors/defs`, and `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` pass. | A separate source-bound partial-coverage disposition records the named source gap without changing the existing command/operation surface. It covers 65 request-schema and 4 path-parameter-alias gaps; the 21 genuinely absent actions remain non-executable. Broader generator verification is serialized behind the concurrent source-lock audit. Green. |
| R2 evidence/origin repair | Red: the R2 audit showed capture metadata gated admission, source-bound credentials could target a caller-selected `base_url`, and operation evidence omitted the V3 Asana lock. | `TestSourceProjectionMutationDispositionCitationIgnoresRetainedArtifactMetadata`, `TestSourceProjectionAcceptsVersion3DocumentProvenance`, `TestSourceBoundOperationRejectsConfiguredOriginBeforeAuthenticationOrIO`, and `TestOperationEvidenceReadsAsanaVersion3DocumentOwnedLock` pass; `operation-evidence --check` is current at 1,774 rows including 249 Asana rows. | Admission uses source identity/URL/location/method/path and typed contract; raw integrity remains a separate lock check. A source-bound operation rejects a different config origin before auth or I/O. Green. |
| Captain 021 capability reconciliation | Red: projection filtered non-mutating GETs by their historical declaration status, so a complete source-backed contract could remain `planned` despite the shared executor. | `TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL`, `TestRetainedAsanaSourceImportRejectsReadProjectionDrift`, and `TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams` pass. The first proves a planned complete direct read becomes an exact source-bound `direct_read`; the second makes `source-import --check` reject an artificially planned direct or fan-out ETL command; the third proves the retained V3 lock. | Projection is capability-based: 106 bounded direct reads and 12 record/pagination-proven streams are implemented, while only `asana.rest.getMembership` retains its concrete `cli-openapi30-reference-sibling-foundation-r1` gap. Green. |
