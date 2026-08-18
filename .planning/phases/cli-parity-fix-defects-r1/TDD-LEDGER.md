# TDD ledger — GitHub command certification defect fixes

| Slice | Red | Green | Refactor / live proof |
| --- | --- | --- | --- |
| Exact integers | Persist a reverse plan with `9007199254740993`; after reload the fake provider must observe that exact path segment, never `9.007199254740992e+15` or the rounded decimal. | Decode interface-backed state numbers without a `float64` hop while preserving existing typed fields. | Exercise at least three real integer-ID commands and assert provider state, then cleanup. |
| Required bodies | Preflight or fake-provider assertions show blob, commit, tree, check-run, branch-protection, and commit-status cannot send their required payload today. | Closed record schemas expose typed flags and the fake provider receives exact required JSON. | Exercise at least three real commands, assert returned/read-back fields, and contain immutable Git objects in the fixture repository. |
| Provider paths | Bundle assertions reject Actions secret/variable endpoints containing `/agents/` and reject the defective project-draft declaration. | Operations, write actions, API surface, and CLI surface agree on provider-served endpoints. | Exercise three corrected Actions commands through `pm`, assert collection/object state, and delete/read back absence. |
| Write accounting | A declared idempotent DELETE returning 404 currently reports `RecordsWritten=1`; command run reports success. | The same response reports zero writes and an incomplete acknowledgement/error; a real successful 2xx still reports one. | Re-run org webhook create/delete plus another affected command with independent reads; never infer mutation from exit status alone. |
| Certification inventory gap | `go test -timeout 20m ./internal/connectors/certify -count=1` reports 606 writes vs 607, 48 primary-user actions vs 49, and 1224 covered endpoints vs 1225 after the invalid user-project REST write was removed. | Assert the exact removed action, the replacement GraphQL command contract, and the REST row's one blocked-duplicate classification; reconcile the pinned write and surface totals to those semantics. | Re-run the previously omitted package explicitly and require PR #4234's `verify` check to pass. |

Red and Green command outputs are appended here before the production change and
after each green slice. A skipped test is not evidence.

## Red checkpoint — 2026-08-18

- Exact integer: `go test -timeout 20m ./internal/app -run '^TestGitHubConnectorCommandPreservesExactLargeIntegerPathAfterPlanReload$' -count=1` failed before the request with `reverse plan command payload changed since approval`; persistence had changed `9007199254740993`.
- Required bodies: `go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubRequiredMutationBodiesReachTheWire$' -count=1` failed for all six commands because the required schema properties/CLI flags were absent.
- Wrong paths: `go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubCorrectedCommandDeclarations$' -count=1` failed on all three Actions aliases and because the user-project command still targeted the live-proven 404 REST route instead of fixed GraphQL.
- False success: `go test -timeout 20m ./internal/connectors/engine -run '^TestWriteDeleteMissingOkStatusDoesNotCountAsWritten$' -count=1` failed because the 404 returned no error and counted `RecordsWritten=1`. The app-level org webhook test has the same red behavior.

## Green checkpoint — 2026-08-18

- Exact integers: state decoding now preserves JSON number lexemes and interpolation renders numeric values in decimal form. `TestGitHubConnectorCommandPreservesExactLargeIntegerPathAfterPlanReload` proves `9007199254740993` survives plan persistence, and `TestInterpolatePathPreservesIntegerIDsWithoutScientificNotation` covers both that exact integer and the recorded `667233046` float regression.
- Required bodies: the six write schemas now expose their provider-required fields, their CLI commands map typed flags into those fields, and `TestGitHubRequiredMutationBodiesReachTheWire` proves the resulting JSON bodies at a provider double.
- Wrong paths: the three mislabelled Agent commands bind to the existing Actions actions and `/actions/` paths. The user-project draft command now builds the closed `addProjectV2DraftIssue` GraphQL input and targets `POST /graphql`; the source-lock REST route is classified as the non-executable duplicate proved by the certification run. `TestGitHubCorrectedCommandDeclarations` and `TestGitHubUserDraftCommandBuildsFixedGraphQLMutation` cover all four declarations and the real command builder.
- False success: allow-listed delete 404s now return `RecordsUnchanged=1`, never `RecordsWritten=1`. Command execution rejects that incomplete mutation acknowledgement, while the closed issue-label cleanup retains explicit already-absent idempotence. The focused engine/app tests and full connector conformance suite pass.

## CI gap red checkpoint — 2026-08-18

- `go test -timeout 20m ./internal/connectors/certify -count=1` reproduced all
  four reported assertions. A base/current inventory comparison proved the
  delta is exactly one action: `projects_create_draft_item_for_authenticated_user`
  is present in the 607-action base and absent from the 606-action branch, with
  no added REST write action. The command itself remains declared as a fixed
  GraphQL direct write, and its invalid source-lock REST endpoint remains
  inventoried as `duplicate`/`blocked`.

## CI gap green checkpoint — 2026-08-18

- The surface inventory now expects 1224 executable covered endpoints plus the
  exact blocked duplicate REST row, rather than calling all 1225 source rows
  executable. Its write coverage is 606.
- The write inventory and sweep now expect 606 actions, explicitly reject
  `projects_create_draft_item_for_authenticated_user`, and pin the resulting
  primary-user prerequisite bucket at 48. The replacement command remains
  covered by `TestGitHubCorrectedCommandDeclarations` through the existing
  `github.graphql.mutation.add-project-v2-draft-issue` operation.
- `go test -timeout 20m ./internal/connectors/certify -count=1` passes.

## Full-verify red checkpoint — 2026-08-18

- GitHub Actions run `32122543487` proved `internal/connectors/certify` green,
  then exposed three stale `cmd/connectorgen` projections of the same transition:
  mutation intents moved from 279 direct writes / 577 reverse ETL writes to
  280 / 576, the API ledger now has one blocked duplicate rather than universal
  executable coverage, and parity completion did not recognize a duplicate
  whose `duplicate_of` target is an implemented fixed operation.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestGitHub(FixtureRequiredMutationCohortGeneratesEveryCandidate|APISurfaceOperationLedgerMetrics|CompleteParityHasNoNonterminalCommandRows)$' -count=1`
  reproduced those three failures locally. The candidate artifact identifies
  the moved command as `projects create-draft-item-for-authenticated-user`, now
  a GraphQL `direct_write` bound to
  `github.graphql.mutation.add-project-v2-draft-issue`.
- Once the intent assertion was reconciled, the same focused test exposed the
  corresponding fixture projection: the retired REST collection cycle moved
  from 489 to 488, while the fixed GraphQL transport's named exception moved
  from 367 to 368. The candidate itself names
  `graphql_transport_not_collection`, proving this is the same command rather
  than an unrelated count change.

## Full-verify green checkpoint — 2026-08-18

- Candidate accounting now asserts the exact user-project command, fixed
  GraphQL operation, GraphQL address, direct-write executor, and named fixture
  exception alongside the 280/576 intent and 488/368 fixture totals.
- API ledger accounting now asserts the exact source REST row as the sole
  low-risk blocked duplicate and subtracts it from executable POST coverage.
- Parity completion treats a duplicate as terminal only when its `duplicate_of`
  string resolves to an operation-backed implemented CLI command; the retired
  REST row resolves specifically to the user-project command.
- `go test -timeout 20m ./cmd/connectorgen -count=1` passes (`98.702s`).
