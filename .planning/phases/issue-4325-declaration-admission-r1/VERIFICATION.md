# Verification — issue 4325 declaration-admission foundation

## Acceptance evidence

- The `connectorgen declaration-admission [dir] [--json]` check requires an
  independent `declaration_admission_inventory.json` plus
  `declaration_admission_sources.json` and `declaration_admissions.json` at the
  definitions root. Inventory selections resolve exact operations from
  connector-owned reviewed source locks; mutable adjacent counts cannot shrink
  the denominator. Its focused tests
  cover one runnable read; deferred reverse-ETL write/delete and binary
  upload/download rows; a retained importer/descriptor gap; missing,
  duplicate, citation-free, malformed, stale, base-path-mismatched,
  lane-changing, destructive-metadata-free, and falsely implemented rows; and
  a complete zero-runnable connector.
- Admission reads the reviewed lock's URL and exact operation identity but
  does not fetch, retain, open, or rehash its provider artifact. A lock may
  carry separate retention metadata; those bytes and hashes are not admission
  prerequisites.
- Authoring admission and the compact shipped ledger share one provider-
  citation canonicalizer. Stored URLs must already use the public-HTTPS
  canonical form; equivalent host case, default-port, query-order, or escaped-
  path spellings cannot split one provider operation into multiple identities.
- Deferred command metadata is projected into the command surface and rejected
  by `commandrunner` with the typed `system/missing_foundation` classification
  before an executor can perform provider I/O.
- A deferred source declaration now requires a bounded missing implementation
  component and non-empty evidence. `blocked_by_default`, lane/method,
  destructive/risk, approval/confirmation, retained source data, and live
  certification are not valid components, so a policy-only block is rejected
  while its command must stay discoverable.
- Source rows own the `none`/`delete`/`destructive` semantic, and declarations
  cannot change it by self-labeling a shared endpoint. GitHub's implemented
  `label delete` action is the destructive green control:
  its source-cited declaration is admitted and the actual commandrunner
  preflight succeeds. Deferred state is therefore endpoint-specific rather
  than a generic delete/destructive classification.
- No connector-owned definition is part of this PR's generic commit range. Existing
  source-lock, surface-sync, runtime-preflight, certification, and live-proof
  gates remain independent and were exercised below where hermetic.
- Captain clarification 007 keeps the separately started Stripe mapping files
  unstaged for `cli-batch1-repair-r1`; they are not part of this PR's commit
  range or certificate claim.

## Audit-repair status

Inbox 008 first repaired DA-008 and DA-009. A later exact-SHA audit of
`683a3c76e` reopened DA-001 through DA-006, DA-010, and DA-011. They are now
green: the denominator is required and counted; source/declaration/runtime
bindings use stable identity; implemented resolution is shared; destructive
semantics are source-owned; GraphQL/shared transports cannot swap operations;
typed missing-foundation survives App/CLI boundaries and oversized metadata;
the compact source ledger is embedded for production-layout preflight; and
source/deferred endpoint identities are structurally validated. The two
unstaged Stripe paths and the concurrently authored Docker Hub connector paths
remain outside this work and are not staged, regenerated, or otherwise used as
certificate evidence.

## Commands and results

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestDeclarationAdmission'` | pass |
| `go test -timeout 20m ./cmd/connectorgen -run '^(TestDeclarationAdmission|TestCheckCLISurfaceEndpointCoverageAllowsDeclarationBoundDeferredCommand)$'` | pass |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestDeclarationAdmissionAdmitsGitHubImplementedDeleteControl$'` | pass |
| `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestPreflightDeferredCommandReturnsNamedFoundationAfterExactTargetValidation$'` | pass |
| `go test -timeout 20m ./internal/connectors/engine -run '^TestCommandSurfaceProjectsDeferredFoundationGap$'` | pass |
| `go test -timeout 20m ./cmd/connectorgen` | pass (153.562s including component/evidence admission regressions) |
| Fresh local project, no credential: `pm github label delete --json` | exit 1, `missing --credential`; command dispatches without provider I/O |
| Fresh local project, no credential: `pm stripe accounts delete --json` | exit 2, `unknown command "accounts delete"`; no generic account-delete projection or provider I/O |
| `go test -timeout 20m ./internal/connectors/commandrunner` | pass |
| `go test -timeout 20m ./internal/connectors` | pass |
| `go test -timeout 20m ./internal/connectors/engine` | pass (13.447s) |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build` | pass |
| `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connectorgen-declaration-admission` | pass |
| `make connectorgen-operation-evidence`, `make github-parity-artifacts-check`, `make connector-boundary`, `make connector-canon-check` | pass |
| `make connectorgen-certification-subject`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, `make connectorgen-certification-sweep` | pass |
| `make release-workflow-check` | pass |
| `go run ./cmd/connectorgen declaration-admission --json` | pass; one required connector/source row and zero findings |
| `go run ./cmd/connectorgen` | expected usage error; confirms the command is listed in the internal generator help |
| `git diff --check` | pass |
| `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestDeclarationAdmission\|TestCheckCLISurfaceEndpointCoverageAllowsDeclarationBoundDeferredCommand)$'` | pass after audit repair |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run '^TestPreflightDeferredCommand'` | pass after audit repair; missing/invalid exact-target resolvers fail before typed `missing_foundation` |
| `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestCommandSurfaceProjectsDeferredFoundationGap$'` | pass after audit repair |
| Focused semantic destructive red/green cases | red showed a POST `destructive_action` could omit metadata; green requires metadata from the exact declared target and rejects a non-destructive POST falsely labelled delete |
| `go test -timeout 20m ./internal/connectors/engine -run '^TestDeferredCommand'` | pass; malformed endpoints, GraphQL staleness, and shared-transport operation swaps fail closed |
| `go test -timeout 20m ./internal/app -run '^TestPlanConnectorCommandPreflightsDeferredCommandBeforeCredentialResolution$'` | pass; missing foundation precedes missing credential |
| `go test -timeout 20m ./internal/cli -run '^TestShippedDeclarationTargetReachesPublicMissingFoundationEnvelopeWithoutAPISurface$'` | pass; `defs.FS` compact ledger and public `internal/missing_foundation` envelope work with no API-surface manifest |
| `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestPreflightDeferredCommand'` | pass; oversized foundation text retains typed `system/missing_foundation` |

### Final clean-checkout audit-repair verification

The complete generic tree immediately before this evidence-only update was
verified from a detached clean worktree. The dirty Stripe and Docker Hub
handoffs were absent from that tree.

| Exact command | Result |
| --- | --- |
| `gofmt -w cmd internal && go mod tidy && git diff --exit-code -- go.mod go.sum && git diff --exit-code` | pass; no formatting, module, or generated working-tree drift |
| `go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors` | pass (`cmd/connectorgen` 299.793s; engine 20.632s; connectors 7.269s) |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner` | pass (32.889s), including all 5,149 commands claimed implemented |
| `go test -count=1 -timeout 20m ./internal/app` | pass (376.835s) |
| `go test -count=1 -timeout 20m ./internal/cli` | pass (636.213s; 83 real local certify invocations) |
| `go vet ./...` | pass |
| `go build -o pm ./cmd/pm` | pass |
| `make tidy-check` | pass |
| `make lint` | pass (`0 issues`) |
| `make docs-check-no-build` | pass (`Validated connector docs in docs/connectors`) |
| `make smoke-no-build` | pass (`smoke ok`) |
| `make agent-contract-check` | pass; canonical contract and projections current |
| `make connectorgen-validate` | pass; 553 connectors, 0 findings |
| `make connectorgen-surface-sync` | pass; 553 connectors, 0 changes |
| `make connectorgen-declaration-admission` | pass; 1 required connector, 1 source operation, 0 findings |
| `make connectorgen-operation-evidence` | pass; 1,525 rows, 5 rollups, fixed-100 passed |
| `make github-parity-artifacts-check` | pass; 17 Node tests and both generated ledgers current |
| `make connectorgen-certification-subject` | pass; current subject current |
| `make connectorgen-certification-matrix` | pass; 3 connector shards current |
| `make connectorgen-certification-candidates` | pass; GitHub candidates current |
| `make connectorgen-certification-sweep` | pass; GitHub sweep current (1,616 rows, 1,612 commands) |
| `make connector-boundary` | pass; 322 files, 553 connectors, 0 findings |
| `make connector-canon-check` | pass |
| `make release-workflow-check` | pass, including pinned dependencies, release parity/tooling/size/layout, and installed GitHub certification proof |
| `scripts/verify-gsd-workflow origin/main` | pass; implementation changes have GSD/TDD evidence |
| Fresh local project: `./pm github label delete --root "$project_root" --json` | exit 1 at `missing --credential`; no provider I/O |

The first clean candidate exposed a real red regression: a literal comparison
between a runtime-relative declaration path and the provider-facing command
endpoint rejected 243 already-runnable commands. The green resolver now proves
the exact stream/write/operation identity while retaining the command
projection's provider endpoint; the full 5,149-command runtime-preflight sweep
and CLI suite pass. The boundary gate also caught a neutral helper spelling
whose normalized text contained the provider token `cal-com`; the helper was
renamed and the whole-tree boundary rerun passed.

The aggregate `go test -timeout 20m ./...` and serial `make verify` are not
run as one process: the repository's `AGENTS.md` explicitly says a per-command
timeout routinely cuts off the full suite and directs agents to run changed
packages plus Make gates individually, leaving the aggregate suite to CI.

Captain instruction 009 also forbids even temporary mutation of the preserved
Stripe handoff. The audit-repair rerun therefore uses the focused generic tests
above plus compile-only repository gates and leaves bundle-loading/generator
checks to the clean committed-tree CI run. The pre-existing Stripe SHA-256
values were rechecked byte-identical after the instruction arrived:
`1f18d5f3cbb4dd4d053af3cdd6505075359b7a58dec1740042686bbea2c2168b`
and `d72234a7c68f8646596579cfbf2a1810cc198a62ce6056116347bbc6bec4183a`.

## CLI/documentation parity

This changes the internal `connectorgen` generator command, not the shipped
`pm` command tree. Its usage string and certification/canon documentation are
updated. `docs/cli/**`, website docs, generated `pm` manual/help, bare `pm`
namespace behavior, and shell completion are not applicable; no files in those
surfaces changed. `make docs-check-no-build` passed.

After captain clarification 007, a later local rerun of
`make docs-check-no-build` reports only that the preserved *unstaged*
`internal/connectors/defs/stripe/cli_surface.json` makes the Stripe manual
stale. This PR must not run the corresponding generator or commit its output;
the mapping lane owns both. The committed generic diff has no Stripe definition
or generated manual change, and `make connector-canon-check` remains green.

## R2 exact-head closure (2026-08-26)

The independent R2 audit's four High findings are closed on clean code SHA
`f97dede0741313bd4661baa1da6c021f568167d8`. Source-operation uniqueness is
now provenance-only while binding uniqueness is an independent invariant;
canonical provider identity and physical transport are retained separately and
related only by closed named proofs; a deferred command must fail the real
implemented commandrunner preflight before its executor-specific missing
foundation can be accepted; and the compact production ledger has exactly one
inventory class with deterministic byte attribution.

This is also the proof required by captain steer 015. Admission completeness is
independent of runnable count: every source-cited operation in any of the six
lanes must have one declared row, while an unavailable operation may remain
discoverable and explicitly `deferred` with its exact missing foundation. A
complete zero-runnable connector therefore passes; omission, false
implementation, or stale deferral fails. Runtime/preflight, credential-bound,
and live certificates remain separate and were not weakened.

| Exact clean-head command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./cmd/connectorgen -run 'TestImplementedCommandEndpointEquivalenceCoversExactFleet\|TestDeclarationAdmissionOutreachRealBundleResolverCompatibility' -v` | pass; 243 non-GraphQL aliases, 4 GraphQL aliases, and synthetic-discovery resolver compatibility over real Outreach stream/write shapes; not shipped CLI or credential proof |
| Four deterministic exact-name shards over `go test ./cmd/connectorgen -list '^(Test\|Example)'` | pass; all 477 tests/examples (`188.823s`, `56.415s`, `90.003s`, `75.841s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/engine` | pass (`19.148s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner` | pass (`30.479s`) |
| `go test -count=1 -timeout 20m ./internal/connectors` | pass (`3.917s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/defs` | pass (`6.480s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/hooks/notion` | pass (`4.395s`) |
| `go test -count=1 -timeout 20m ./internal/app` | pass (`318.334s`) |
| `go test -count=1 -timeout 20m ./internal/cli` | pass (`575.419s`) |
| `go vet ./...`; `make lint` | pass; lint reports 0 issues |
| `go build -o ./pm ./cmd/pm`; `make tidy-check`; `make docs-check-no-build`; `make smoke-no-build` | pass; exact-head binary, module state, connector docs, and ETL/reverse-ETL smoke |
| `make connectorgen-declaration-admission`; `make connectorgen-surface-sync` | pass; 1 connector/source row with 0 findings, and 553 connectors with 0 drift |
| `make connector-boundary` | pass; 322 files, 553 connectors, 0 findings |
| `make release-workflow-check` | pass; pinned dependencies, target parity, tooling, size/layout, and installed GitHub archive proof |
| `scripts/verify-gsd-workflow origin/main` | pass; implementation changes have GSD/TDD evidence |
| `git diff --exit-code`; `git status --short` | pass; detached exact-head verification tree clean |

The unsplit exact-head `go test -count=1 -timeout 20m ./cmd/connectorgen`
attempt was externally terminated with exit 143 after more than three minutes
and no test failure, before Go's timeout. The complete 477-name inventory was
therefore partitioned by list index modulo four and every shard passed. The
aggregate `go test ./...` and serial `make verify` remain intentionally delegated
to CI under the repository's explicit per-command-timeout guidance; all
applicable constituent gates were run locally.

The boundary analyzer initially flagged the neutral local name
`canonicalComparable` because normalized text contained the provider token
`cal-com`. It was renamed to `declaredComparable`; no exception was added, and
the exact-head whole-tree boundary rerun passed.

## R3 canonical-citation repair verification (2026-08-26)

The repair checkpoint `1d3ac8d273235664c92a84c170d1946ce56a3339` was
verified from a clean clone inside the disposable task worktree. The preserved
Stripe and Docker Hub handoffs were absent from that snapshot and remain
unstaged in the worker tree. A preliminary broad dirty-tree package run failed
only because the preserved Stripe projection lacks its concurrent
`foundation_gap.component`; that known handoff failure was not rerun or
modified. The exact committed snapshot is green.

| Exact command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./internal/safety ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors ./internal/connectors/defs` | pass (`1.044s`, `18.469s`, `33.958s`, `7.423s`, `10.266s`) |
| `go test -count=1 -timeout 20m ./cmd/connectorgen` | pass (`280.359s`), including canonical citation adversarial cases, the 243+4 fleet census, and truthfully named Outreach resolver compatibility |
| `go test -count=1 -timeout 20m ./internal/app` | pass (`337.211s`) |
| `go test -count=1 -timeout 20m ./internal/cli` | pass (`636.015s`) |
| `go vet ./...`; `go build -o ./pm ./cmd/pm`; `make tidy-check`; `make lint` | pass; lint includes `internal/safety` and reports 0 issues |
| `make docs-check-no-build`; `make smoke-no-build`; `make agent-contract-check` | pass; connector docs validate, local-only smoke succeeds, canonical contract is current |
| `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connectorgen-declaration-admission` | pass; 553 bundles / 0 findings, 0 surface drift, 1 connector / 1 source row / 0 admission findings |
| `make connectorgen-operation-evidence`; `make github-parity-artifacts-check` | pass; 1,525 operation-evidence rows and 17 GitHub parity tests/artifacts current |
| `make connectorgen-certification-subject`; `make connectorgen-certification-matrix`; `make connectorgen-certification-candidates`; `make connectorgen-certification-sweep` | pass; subject/matrix/candidates current and GitHub sweep current at 1,616 rows / 1,612 CLI commands |
| `make connector-boundary`; `make connector-canon-check`; `make release-workflow-check` | pass; 323 files / 553 connectors / 0 findings, canon current, all release/archive checks pass |
| `scripts/verify-gsd-workflow origin/main`; exact-clone `git status --short` | pass; GSD/TDD evidence present and no tracked verification drift |
| Post-checkpoint review fix: `go test -count=1 -timeout 20m ./internal/safety`; focused authoring/ledger tests; scoped `golangci-lint` | pass; plain and percent-escaped dot segments now fail closed, 0 lint issues |

The aggregate `go test ./...` and serial `make verify` remain assigned to CI by
the repository's per-command-timeout guidance; every applicable constituent
gate was run locally. No provider I/O, credentialed check, write/delete, source
lock/hash prerequisite, or live certification was performed. This foundation
still does not prove shipped Outreach commands: final merge validation requires
the real combined-head Outreach mapping/pilot after #4350 repair with committed
CLI/source evidence, credential-boundary and zero-transport proof, and a fresh
independent audit.

## R5 reviewed-source inventory and public-input repair (2026-08-26)

The R5 repair binds admission to one exact operation selected from a
connector-owned reviewed source lock. Provider provenance owns
source-operation uniqueness; runtime binding uniqueness remains separate. The
independent inventory is the denominator, so deleting adjacent catalog rows or
restoring mutable count fields cannot produce a false pass.
`unsupported_with_provider_evidence` is discoverable and denominator-visible
in all six lanes without claiming an executor or missing foundation.

Public connector input validation now shares commandrunner's real flag rules
and runs after help but before App/credential construction. The first clean
full CLI run exposed one category regression: unknown command paths became
validation errors. Its existing controls were red, the wrapper was narrowed,
and both the focused controls and full package rerun are green.

| Exact clean-head command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestDeclarationAdmission'` | pass |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run 'TestPreflightProviderEvidencedUnsupportedReturnsTypedTerminal\|TestPreflightRequestValidatesDeclaredInputsWithoutExecutor'` | pass |
| `go test -count=1 -timeout 20m ./internal/connectors/engine -run 'TestDeclarationTargetLedger\|TestCommandSurfaceProjectsDeferredFoundationGap'` | pass |
| `go test -count=1 -timeout 20m ./internal/cli -run 'TestGitHubLabelDeleteValidatesRequiredInputBeforeCredentialResolution\|TestConnectorCommandInputDefectsFailBeforeWithApp'` | pass |
| `go test -count=1 -timeout 20m ./cmd/connectorgen` | pass (`296.568s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner` | pass (`31.489s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/engine` | pass (`16.322s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/defs` | pass (`6.291s`) |
| First `go test -count=1 -timeout 20m ./internal/cli` | red (`745.861s`): four existing unknown-path controls exposed validation/usage drift |
| Focused rerun of those four controls | pass (`20.528s`) |
| Final `go test -count=1 -timeout 20m ./internal/cli` | pass (`589.430s`) |
| `gofmt -w cmd internal`; `go mod tidy`; scoped `go vet`; `go build ./cmd/pm` | pass; no source/module drift |
| `make docs-check-no-build`; `make smoke-no-build`; `make lint` | pass; docs valid, smoke succeeds, lint reports 0 issues |
| `make agent-contract-check`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connectorgen-declaration-admission` | pass; 553 bundles / 0 findings, 0 surface drift, 1 connector / 1 selected operation / 0 admission findings |
| `make connectorgen-operation-evidence`; `make github-parity-artifacts-check` | pass; 1,525 evidence rows and 17 GitHub ledger tests/artifacts current |
| `make connectorgen-certification-subject` | initially stale after declaration changes; regenerated only `current-subject.json` from the clean tree |
| `make connectorgen-certification-matrix`; `make connectorgen-certification-candidates`; `make connectorgen-certification-sweep` | pass; downstream artifacts current |
| `make connector-boundary`; `make connector-runtime-preflight`; `make connector-canon-check` | pass; 323 files / 553 connectors / 0 findings, every implemented command passes real preflight, canon current |
| `make release-workflow-check` | pass, including installed GitHub certification archive proof |

The aggregate `go test ./...`, aggregate `make verify`, and repository-wide
`go vet ./...` were not launched in this disk-floor window. Repository guidance
assigns the aggregate suite to CI and requires changed packages and constituent
gates separately; Firstmate instruction 020 additionally required serialization
and preference for focused checks while only 16 GiB remained. No provider I/O,
credential, write/delete execution, live certification, bulk regeneration,
Stripe definition, or Docker Hub definition was used.

## R6 declaration/retention boundary and effective-config verification (2026-08-26)

The exact R6 generic diff was validated from the clean nested checkout rooted
at audited head `ab2c5e3933e0dc1355948d3585b269c46f75754d`. Only the generic
foundation paths were applied there; the preserved Stripe and Docker Hub
handoffs were absent. Admission now consumes strict mapping evidence without
consuming retention evidence, while source import still rejects the same
missing or malformed byte/hash fields. Effective `config.*` values are fully
coerced before App/credentials, explicit argv wins, and public errors name only
the declaration target and flag. Inventory schema v2 is mandatory.

| Exact clean-checkout command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestDeclarationAdmission'` | pass, including legacy/v3 retention independence, strict mapping identity, inventory v2, and legacy-v1 rejection |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run 'TestPreflightRequestValidatesConfiguredDeclaredValuesWithoutExecutor\|TestPreflightRequestValidatesDeclaredInputsWithoutExecutor'` | pass for enum/type/format/empty/byte/range/cardinality, redaction, and argv precedence |
| `go test -count=1 -timeout 20m ./internal/cli -run '^TestFreshchatConfiguredEnumValidatesBeforeCredentialResolution$'` | pass using the shipped Freshchat bundle and isolated registry |
| `go test -count=1 -timeout 20m ./cmd/connectorgen` | pass (`330.178s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/commandrunner` | pass (`34.497s`) |
| `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/defs` | pass (`19.286s`, `7.255s`) |
| `go test -count=1 -timeout 20m ./internal/cli` | pass (`608.975s`) |
| Built `pm`; invalid `freshchat agents list --config agents_sort_order=sideways --json` | exit 3 validation before credentials; output is redacted |
| Same invocation plus `--sort-order asc` | argv override succeeds through input preflight and reaches exit 1 `missing --credential`; no provider call |
| `make tidy-check`; `go vet ./...`; `make lint`; `make agent-contract-check`; `make build` | pass; module files unchanged, lint reports 0 issues, canonical contract current |
| `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connectorgen-declaration-admission` | pass; 553 bundles / 0 findings, 0 surface drift, 1 connector / 1 selected operation / 0 admission findings |
| `make connectorgen-operation-evidence`; `make github-parity-artifacts-check` | pass; 1,525 evidence rows and all 17 GitHub ledger tests/artifacts current |
| `make connectorgen-certification-subject` | initially reported required declaration fingerprint drift; regenerated only `internal/connectors/certifications/current-subject.json`, then passed |
| `make connectorgen-certification-matrix`; `make connectorgen-certification-candidates`; `make connectorgen-certification-sweep` | pass; downstream artifacts current, GitHub sweep 1,616 rows / 1,612 commands |
| `make docs-check-no-build`; `make smoke-no-build` | pass; connector docs validate and local ETL/reverse-ETL smoke succeeds |
| `make connector-boundary connector-canon-check` | pass; 323 files / 553 connectors / 0 findings and canon current |
| `make release-workflow-check` | pass, including installed GitHub certification archive proof |

The aggregate `go test -timeout 20m ./...` and serial `make verify` remain
assigned to CI under the repository's explicit per-command-timeout guidance;
all affected packages and every applicable non-aggregate constituent gate ran
locally. No provider I/O, credential use, live proof, provider write/delete,
retained artifact read/hash, or connector-owned definition change was used.
