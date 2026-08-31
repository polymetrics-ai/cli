# Verification — Issue #4293 retained source-evidence receipt cohort R1

## Pending checklist

- [x] Red focused test recorded before implementation.
- [x] Green focused cohort/sidecar tests recorded after implementation: exit 0 in 88.236s.
- [x] `go run ./cmd/connectorgen source-operation-mapping-cohort <cohort> --check --check-retention-receipts` succeeds and reports only mapping-only receipt facts: 10/4,341 base rows then 8/2,340/0 retention receipts.
- [x] `go run ./cmd/connectorgen source-operation-mapping-cohort --help` and root `--help` expose the check-only flag.
- [x] JSON parse, `gofmt`, focused `go vet`, `go run ./cmd/agentcontractgen check`, and `git diff --check` are green.
- [x] Diff contains no source lock, lane matrix, connector runtime artifact, source-import/materialize/projection, engine, certification, Atlas, or receiver change.
- [ ] Candidate commit/push is reported for independent review only; no parent integration or main merge is performed.

## Foundation disposition

No runtime foundation is requested. The receipt validator is authoring-only
evidence binding. It cannot make a source cell executable; subsequent runtime
materialization still requires an exact source binding, a closed artifact, and
a credential-bound execution witness.

## Scoped full-package baseline observation

`GOCACHE=/private/tmp/gocache-4293-retained-source-receipt-r1 go test -timeout
25m ./cmd/connectorgen -count=1` exited `1` after 510.095s. The failures are
outside this slice's changed paths and concern existing Asana/GitLab projection
and ledger expectations: `TestEnabledConnectorContractsKeepExecutableLanesImplementedWhenSourceMappingIsPartial`,
`TestOperationEvidenceGitLabSourceLockBridge`,
`TestRetainedAsanaSourceImportRejectsReadProjectionDrift`,
`TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation`,
`TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant`, and
`TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical`.

The same exact six-test subset was rerun from a clean detached worktree at the
frozen base `ceaae873aef0dd19aa23c036b9cb598f9b3eacc8`, with
`GOCACHE=/private/tmp/gocache-4293-baseline-ceaae` and `-count=1`. It also
exited `1` (13.834s), establishing the failures as pre-existing at the frozen
base:

- `enabledcontract_final_test.go`: Asana ETL coverage is already `complete`
  rather than the historical `partial` expectation, and GitLab binary download
  is already `mapped_unproven` rather than a deferred foundation.
- `operationevidence_test.go`: GitLab already reports 967 runtime-enabled
  identities rather than the historical 733 expectation.
- `sourceprojection_test.go`: the Asana retained read-projection check already
  detects descriptor/bundle drift; the mutation-disposition test already finds
  unresolved Asana actions/gaps; the isolated action-variant fixture already
  produces four CLI changes rather than one; and the installed GitHub
  projection already reports 109 operation and 1,086 CLI changes.

The tests themselves live only in
`cmd/connectorgen/enabledcontract_final_test.go`,
`cmd/connectorgen/operationevidence_test.go`, and
`cmd/connectorgen/sourceprojection_test.go`. #4293 changes none of those
files, nor any Asana, GitLab, GitHub, source-import, source-materialize, or
source-projection path. The focused receipt suite is green; no unrelated repair
is authorized in #4293.

## Static validation

All commands below ran from this candidate with its isolated cache unless noted
otherwise:

```text
gofmt -d cmd/connectorgen/main.go cmd/connectorgen/retainedsourcemapping.go \
  cmd/connectorgen/sourceoperationmappingcohort.go \
  cmd/connectorgen/sourceoperationmapping_test.go
# no output

jq empty data/connector-canon/batch1-source-operation-mapping-cohort.json \
  internal/connectors/defs/{bitbucket,circleci,dockerhub,jira,notion,sentry,stripe,vercel}/sources/*-retained-mapping-contract.json
# exit 0

GOCACHE=/private/tmp/gocache-4293-retained-source-receipt-r1 \
  go vet ./cmd/connectorgen
# exit 0

GOCACHE=/private/tmp/gocache-4293-retained-source-receipt-r1 \
  go run ./cmd/agentcontractgen check
# agentcontractgen: canonical contract and registered projections are current

git diff --check
# no output
```

The candidate diff is limited to the cohort/help/builder/test implementation
and its issue-local plan, TDD ledger, and verification record. It has no
connector lock, matrix, definition artifact, engine, source import,
materialization, projection, certification, Atlas, or receiver path.
