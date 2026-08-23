# Shared Test Fixture Evidence — 2026-08-24

Scope: evidence requested for the open `verify-shared-test-fixtures` decision.
No shared test or production code was changed for this report.

Branch head examined: `7fcb4841a68220e0cfa69b1821629e13947e9c5d`.
Comparison parent: `8127de41875ccdf992abaf47789eefbaeb522c32` (the merge-base with the
then-current `origin/main`).

## Exact reproductions

### `cmd/connectorgen/operationevidence_test.go`

Command:

```sh
go test -timeout 20m -count=1 -run '^(TestOperationEvidenceFixed100RejectsEveryRegression|TestOperationEvidenceCheckRunsFixed100Gate)$' ./cmd/connectorgen
```

Result (exit 1):

```text
--- FAIL: TestOperationEvidenceFixed100RejectsEveryRegression (0.86s)
    operationevidence_test.go:232: unmodified fixed cohort rejected: asana.rest.getCustomFieldsForWorkspace is absent from operation evidence
--- FAIL: TestOperationEvidenceCheckRunsFixed100Gate (1.42s)
    operationevidence_test.go:262: fixed-100 check exit=1 stderr="connectorgen operation-evidence: fixed-100 validation failed: asana.rest.getCustomFieldsForWorkspace is absent from operation evidence\n"
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	3.379s
FAIL
```

The requested source lines are:

```go
226 artifact, _, _ := runOperationEvidenceForTest(t, root, "")
227 fixed := loadOperationEvidenceFixed100(t, root)
228 if len(fixed.Rows) != 100 {
229     t.Fatalf("fixed cohort rows = %d, want 100", len(fixed.Rows))
230 }
231 if err := validateOperationEvidenceFixed100(artifact, fixed); err != nil {
232     t.Fatalf("unmodified fixed cohort rejected: %v", err)
233 }
...
273 func operationEvidenceWorkspace(t *testing.T) string {
274     t.Helper()
275     root := t.TempDir()
276     copyOperationEvidenceTree(t, repositoryRoot(t), root, "internal/connectors/defs", "github")
277     copyOperationEvidenceFile(t, repositoryRoot(t), root, "internal/connectors/operation-evidence-fixed-100.json")
278     copyOperationEvidenceFile(t, repositoryRoot(t), root, "internal/connectors/certifications/current-subject.json")
```

Fixture expectation versus branch result:

| Item | Fixture expects | Branch produces |
| --- | --- | --- |
| definition workspace | only `internal/connectors/defs/github` is copied | generated `operation-evidence.json` contains 1,525 GitHub rows |
| fixed cohort input | current repository's 100-row `operation-evidence-fixed-100.json` | current cohort selects 37 GitHub, 33 Asana, 25 Docker Hub, 3 Jira, and one each of Bitbucket and CircleCI |
| validation | every selected row appears in the GitHub-only generated artifact | fails at the first selected Asana row: `asana.rest.getCustomFieldsForWorkspace` |

The current operation evidence has **5,903** rows; the comparison parent has
**1,525** rows. The difference is **4,378**, exactly the batch-one declared
operation denominator. The temporary fixture, however, intentionally generates
only the GitHub subset, so it cannot contain a newly selected non-GitHub fixed
cohort row.

### `internal/app/transport_composition_test.go`

Command:

```sh
go test -timeout 20m -count=1 -run '^TestDefinitionTransportFactoriesSelectDeclaredEvidence$' ./internal/app
```

Result (exit 1):

```text
--- FAIL: TestDefinitionTransportFactoriesSelectDeclaredEvidence (1.43s)
    transport_composition_test.go:297: source factory evidence = &synctransport.DefinitionFactory{
      Reference: connectors.TransportExecutorReference{Family:"declarative_api", ID:"declarative_stream_source"},
      SourceEvidence: connectors.ConformanceEvidenceReference{
        Suite:"batch1_declarative_stream_transport",
        RunID:"asana_all_executable_streams_v1"
      },
      AcceptedSourceEvidences:[]connectors.ConformanceEvidenceReference{
        {Suite:"batch1_declarative_stream_transport", RunID:"bitbucket_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"circleci_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"dockerhub_all_executable_streams_v1"},
        {Suite:"declarative_stream_transport", RunID:"all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"gitlab_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"jira_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"notion_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"sentry_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"stripe_all_executable_streams_v1"},
        {Suite:"batch1_declarative_stream_transport", RunID:"vercel_all_executable_streams_v1"}
      },
      DestinationEvidence: empty,
      AcceptedDestinationEvidences:nil,
      BuildSource:(...), BuildDestination:nil
    }, want declaration connectors.ConformanceEvidenceReference{
      Suite:"declarative_stream_transport", RunID:"all_executable_streams_v1"
    }
FAIL
FAIL	polymetrics.ai/internal/app	2.527s
FAIL
```

The exact assertion at line 297 is:

```go
286 var sourceFactory, destinationFactory *synctransport.DefinitionFactory
287 for index := range factories {
288     factory := &factories[index]
289     if factory.Reference == source.Executor && factory.BuildSource != nil {
290         sourceFactory = factory
...
296 if sourceFactory == nil || sourceFactory.SourceEvidence != source.Conformance {
297     t.Fatalf("source factory evidence = %#v, want declaration %#v", sourceFactory, source.Conformance)
298 }
```

Fixture expectation versus branch result:

| Item | Fixture expects | Branch produces |
| --- | --- | --- |
| inspected declaration | GitHub source conformance: `declarative_stream_transport/all_executable_streams_v1` | one registry-wide `declarative_api/declarative_stream_source` factory |
| factory primary evidence | GitHub's generic evidence record | alphabetical first shared declaration, Asana: `batch1_declarative_stream_transport/asana_all_executable_streams_v1` |
| accepted evidence | assertion does not inspect acceptance set | 10 accepted records, including the expected GitHub generic record and the nine other batch-one source records; primary Asana makes 11 declared source records total |

## Fixed-100 cohort integrity check

The fixed cohort remains exactly **100** rows.

| Check | Measured result |
| --- | --- |
| parent fixed cohort | 100 GitHub source IDs |
| branch fixed cohort | 100 source IDs: 37 GitHub, 33 Asana, 25 Docker Hub, 3 Jira, 1 Bitbucket, 1 CircleCI |
| original cohort IDs represented in current generated operation evidence | 100 of 100 |
| semantic projection mismatch for those original 100 IDs | 0 |
| raw ordering-only differences | 2 classifications: `github.rest.pulls/get` and `github.rest.repos/get-release-asset`; the same two values appear in a different order (`direct_read`, `binary_download`) |
| parent IDs no longer selected by the current fixed cohort | 63 of 100; they are replaced in the **selection fixture**, not removed from operation evidence |
| current fixed cohort IDs absent from current operation evidence | 0 |

The parent and current-main copies of the fixed fixture have the same SHA-256:
`1c7a80a30adb1da64dec54fa35edd1bbc90d1f1db04a68a5554de3fe94ffdddc`.
The branch copy SHA-256 is
`74c5c8804f660e14f0abbb836ddae463e4476fb3d16f15011e9b8e488c7017ad`.

## Evidence answer to the decision question

The observed delta is an **addition and selection-fixture change**, not evidence
that any of the original guarded GitHub operation surfaces was lost:

- the branch adds 4,378 generated evidence rows (1,525 to 5,903);
- all 100 original fixed source IDs remain in the branch operation evidence;
- their normalized semantic projections have zero mismatches; and
- the failing temporary workspace copies only GitHub while loading the branch's
  cross-connector selection cohort.

For the transport assertion, the expected GitHub conformance record remains in
the factory's accepted evidence set. The asserted `SourceEvidence` field now
holds the first registry-wide shared source declaration (Asana), rather than a
GitHub-specific primary value.

This report initially did not decide whether the temporary fixture should be
regenerated or its guard altered. Firstmate subsequently authorized the bounded
repair below while explicitly preserving the existing 100-row cohort.

## Authorized repair result

Only the two shared tests and issue planning evidence changed. The fixed
fixture, connector definitions, source locks, generated artifacts, and
production transport composition remain unchanged.

- The temporary operation-evidence workspace now reads its fixed reference,
  derives the six connector prefixes (Asana, Bitbucket, CircleCI, Docker Hub,
  GitHub, and Jira), and copies only those definition trees and their generated
  website rows.
- `TestOperationEvidenceFixed100RejectsEveryRegression` removes
  `github.rest.issues/list-for-repo` only from a separate `t.TempDir` source
  lock and requires the direct fixed-cohort validator to reject that exact ID.
- `TestOperationEvidenceCheckRunsFixed100Gate` performs the same temporary
  removal and requires the CLI `operation-evidence --check` path to reject the
  exact ID. This proves both pre-existing fixed-cohort guard paths still fail
  on a real GitHub source-row loss; `t.TempDir` removes the altered copy after
  each test.
- `TestDefinitionTransportFactoriesSelectDeclaredEvidence` now requires the
  exact GitHub conformance reference in the shared factory's primary-or-
  accepted set. It does not accept an arbitrary evidence record, and it retains
  the existing exact GitHub destination-evidence assertion.

Focused green results:

```text
ok  polymetrics.ai/cmd/connectorgen  16.514s
ok  polymetrics.ai/internal/app      2.896s
```

An attempted full `cmd/connectorgen` package run was blocked by the local
filesystem: an unrelated Freshservice fixture setup reported `no space left on
device`. `df -h` showed only 6.5 GiB available on the shared data volume. No
temporary directory was deleted because it is outside this worktree and may be
owned by another task.

## Commands and data checks used

```sh
go test -timeout 20m -count=1 -run '^(TestOperationEvidenceFixed100RejectsEveryRegression|TestOperationEvidenceCheckRunsFixed100Gate)$' ./cmd/connectorgen
go test -timeout 20m -count=1 -run '^TestDefinitionTransportFactoriesSelectDeclaredEvidence$' ./internal/app
git merge-base HEAD origin/main
git show 8127de41875ccdf992abaf47789eefbaeb522c32:internal/connectors/operation-evidence-fixed-100.json
git show 8127de41875ccdf992abaf47789eefbaeb522c32:internal/connectors/operation-evidence.json
sha256sum internal/connectors/operation-evidence-fixed-100.json
```
