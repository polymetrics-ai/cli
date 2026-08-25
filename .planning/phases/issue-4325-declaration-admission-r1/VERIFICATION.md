# Verification — issue 4325 declaration-admission foundation

## Acceptance evidence

- The new `connectorgen declaration-admission [dir] [--json]` check requires
  `declaration_admission_sources.json` and `declaration_admissions.json` at the
  definitions root. Nonzero expected counts make missing catalogs, omitted
  rows, and zero-work runs fail. Its focused tests
  cover one runnable read; deferred reverse-ETL write/delete and binary
  upload/download rows; a retained importer/descriptor gap; missing,
  duplicate, citation-free, malformed, stale, base-path-mismatched,
  lane-changing, destructive-metadata-free, and falsely implemented rows; and
  a complete zero-runnable connector.
- The on-disk fixture has only source URL, exact citation, stable source and
  binding identities, endpoint, lane, command, destructive semantic, and
  state. Its raw provider operation ID is empty, and it contains neither an
  artifact nor a hash, proving none is an admission prerequisite.
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
