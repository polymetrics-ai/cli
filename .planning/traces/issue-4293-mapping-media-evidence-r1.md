# Issue 4293 — Source-cited binary-media evidence repair R1

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- Base branch: `codex/4293-mapping-controls-r1` at immutable commit `82445ed8dc445bbf0b34b3a9423588f4e9a4b0fa`.
- Merges into: `codex/4293-mapping-controls-r1` → Batch R1 parent branch → `main`.
- Delivery: A scoped commit is pushed for independent review; it is not merge-ready until that review accepts the repair.
- Working branch: `codex/4293-mapping-media-evidence-r1`.
- Task: Remove the closed binary MIME allow-list from source-operation mapping admission. Accept an exact-source-cited, concrete non-JSON provider media type such as Vercel's `application/gzip` without adding a gzip special case or a replacement allow-list. Keep explicit rejection when the evidence is uncited, malformed, wildcarded, or contradicts a JSON media assertion.
- Verification: Focused red/green/refactor tests in `./cmd/connectorgen`; mapping CLI check using the focused fixture; `gofmt`; `go vet ./cmd/connectorgen`; `go run ./cmd/agentcontractgen check`; `git diff --check`; JSON validity of any changed JSON file (none expected).

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Vercel-shaped `application/gzip` evidence admits a binary-upload lane | live | The manifest fixture changes a cited request media fact to `application/gzip` and an applicable binary-upload cell; the actual mapping checker exits zero. |
| Parameterized media is normalized | live | `application/gzip; charset=binary` passes through the same checker, proving media parameters do not recreate a closed MIME list. |
| A non-allowlisted source-cited binary media type is admitted | live | `application/x-provider-archive` passes with a cited applicable binary lane, without a provider- or MIME-specific code branch. |
| Mapping evidence cannot be fabricated or contradicted | live | A wrong-node citation and a concrete JSON or malformed/wildcard media assertion with an applicable binary lane produce named checker findings. |

## Discussion / design record

- `source_operation_mapping` is an authoring-only mapping control. It must not change source locks, connector definitions, source import, runtime execution, transport, certification, credentials, or Foundation Atlas.
- The manifest already binds every fact citation to the exact reviewed source operation URL and source-lock location. That exact citation is the provenance boundary for media facts.
- Media classification must normalize media parameters with the standard library and reject malformed or wildcarded values. JSON media (`application/json` and `application/*+json`) remains contradictory to a binary claim. Other concrete source-cited media types are provider evidence; they are not gated by a centrally maintained content-type allow-list.
- The runner has no compatible isolated GSD role and repository policy forbids role spawning. The generated GSD prompts are therefore executed inline and recorded here.

## Plan (TDD)

1. **Red:** extend `TestSourceOperationMapping...` with source-cited `application/gzip`, parameterized `application/gzip`, and a non-allowlisted concrete provider type; each must fail against the closed allow-list. Add negative malformed/wildcard and JSON contradiction cases plus wrong-node citation coverage.
2. **Green:** replace the allow-list helper with a small concrete-media normalizer that relies on the already exact-bound source citation and rejects only non-concrete, malformed, or JSON media claims.
3. **Refactor:** keep diagnostics deterministic, run formatting and focused verification, and confirm the changed package has no runtime or connector-definition dependency.

## Required skills and lifecycle

Loaded: repository required-skills routing, `go-engineering` (fundamentals), `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

Resolved GSD path: `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`, with their generated prompts inspected. This narrow review-fix follows the same issue #4293 phase and uses the inline/manual fallback above.

## TDD ledger

### Red

Before changing production code, the new acceptance cases failed against the
closed allow-list:

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'TestSourceOperationMappingCheck(AcceptsSourceCitedConcreteBinaryMedia|RejectsUncitedOrContradictoryBinaryMedia)$' \
  -count=1
```

The three source-cited cases (`application/gzip`, parameterized gzip, and
`application/x-provider-archive`) each failed with
`binary_upload applicability is not supported by request media evidence`.
This was the expected MAP-R1-004 regression. The negative cases were added to
ensure the repair does not turn an uncited, JSON, malformed, or wildcarded
claim into evidence.

### Green

The mapping helper now uses `mime.ParseMediaType`, rejects malformed/wildcarded
or JSON media, and accepts any other concrete source-cited provider type. No
provider name or media literal is hard-coded.

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'TestSourceOperationMappingCheck(AcceptsSourceCitedConcreteBinaryMedia|RejectsUncitedOrContradictoryBinaryMedia)$' \
  -count=1
# ok  polymetrics.ai/cmd/connectorgen  6.266s

go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestBatch1SourceOperationMappingCohort|TestSourceOperationMapping)' \
  -count=1
# ok  polymetrics.ai/cmd/connectorgen  43.162s

go test -timeout 20m ./internal/connectors/engine -count=1
# ok  polymetrics.ai/internal/connectors/engine  25.499s

go vet ./cmd/connectorgen ./internal/connectors/engine
# exit 0

go run ./cmd/connectorgen source-operation-mapping-cohort \
  data/connector-canon/batch1-source-operation-mapping-cohort.json --check
# connectorgen source-operation-mapping-cohort: 10 connector(s), 4341 source operation(s), 0 finding(s)

go run ./cmd/connectorgen declaration-admission
# connectorgen declaration-admission: 1 connector(s), 1 source operation(s), 0 finding(s)

go run ./cmd/agentcontractgen check
# agentcontractgen: canonical contract and registered projections are current
```

The full changed-package check was also run:

```text
go test -timeout 20m ./cmd/connectorgen -count=1
```

It exited 1 after 863.948s in six known Batch R1 baseline tests:
`TestImplementedCommandEndpointEquivalenceCoversExactFleet`,
`TestOperationEvidenceGitLabSourceLockBridge`,
`TestRetainedAsanaSourceImportRejectsReadProjectionDrift`,
`TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation`,
`TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant`, and
`TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical`.
They are the same six failures recorded before this change in
`.planning/phases/issue-4293-source-operation-multilane-manifest-r1/VERIFICATION.md`;
this diff touches none of their source-projection or connector-artifact paths.

### Refactor / review

`gofmt` was run on both changed Go files. `git diff --check` and the final
scoped-diff review are pending immediately before commit. Independent Codex
review is required after the scoped commit is pushed; this trace does not
declare merge readiness.
