# VERIFICATION — issue #4368 zero-operation source-reference foundation

## Status

Foundation implementation is green locally. Firstmate authorized publication as
a scoped foundation-only PR on 2026-08-27. Cohort fan-in remains an explicit
source-evidence gap because the checked-in retained bytes and artifact manifests
are absent; no live provider fetch or synthetic artifact may replace them.

## Acceptance checklist

- [x] Explicit zero-operation rendered coverage validates only with retained bytes, source lock, manifest provenance and a closed marker.
- [x] Missing, malformed, unverified, accidental-empty, duplicate and mixed-document variants fail closed with location.
- [x] Non-empty rendered-reference and OpenAPI v1-v3 source-lock/import behavior has unchanged byte/count checks.
- [ ] Exactly 720 source-cited deferred rows reconcile as 187 Amplitude + 49 Dremio + 193 Ashby + 84 Workable + 207 HiBob. Explicit source-evidence gap: the current Batch 8–10 branch has only locks/crosswalks/dispositions, no `sources/artifacts/` files or retained-artifact manifests for these five source locks. Firstmate authorized this foundation-only PR without claiming that exercise.
- [ ] Rows preserve exact source citation, provider operation ID, stable identity, applicable lane, and exactly one `missing_foundation` disposition.
- [x] Deferred-command preflight guard remains typed and pre-I/O: targeted commandrunner tests cover named-foundation, invalid-target, unavailable-disposition, and every implemented command's real runtime preflight. No cohort command exists in this scoped PR to exercise or reclassify.
- [x] Generator/projection/evidence/surface-sync/validate and bounded JSON duplicate invariants pass, except the pre-existing GitHub `source-import --check` descriptor drift recorded below.
- [x] Formatting, vet, build, diff, individual repository gates, and exact-head local review pass. Rebase, PR API base verification, CI, and independent exact-head Codex audit remain publication/review steps.

## Foundation evidence

- RED: `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportV3RenderedReferenceCoverageOnlyRequiresExplicitClosedProof$' -count=1` failed before implementation with `json: unknown field "coverage_only"`.
- GREEN: `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportV3RenderedReferenceCoverageOnly' -count=1` passed.
- Regression: `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportVersion3RenderedReference|TestSourceRetainRetainsRenderedReferenceAndBundleArtifacts|TestSourceImport_RejectsUnknownSectionAndIndependentIndexOverflow)$' -count=1` passed.
- Full-package correction: `go test -timeout 20m ./cmd/connectorgen -count=1` first failed only because the existing `TestSourceImportVersion3RenderedReferenceRetainsCapturedEvidenceWithoutOperations` fixture omitted the newly required discriminator. The fixture now declares `coverage_only:true`; its focused regression with all coverage-only tests passes. It passed in 193.136 seconds, then again at exact code head in 315.697 seconds after the local-only command/projection witness was added.
- Batch source check: `git ls-tree -r --name-only fm/cli-map-batch8910-r1 -- internal/connectors/defs` shows only the five lock/crosswalk/disposition files, not retained artifact content or manifests. The five lock inventories total 720 operations and list 15 zero-operation rendered documents; their 15 pinned digest paths are absent from `git rev-list --all --objects`.

## Local-only importer/projection witness

`TestSourceImportV3RenderedReferenceCoverageOnlyRetainsAndVerifiesLockedBytes`
first retains fixture bytes with a fake setup fetcher. It then invokes the real
`source-import` command and `--check` with a nil fetcher, so production must
construct `newConnectorSourceImportRetainedArtifactFetcher` and read only the
checked-in lock, manifest, and content-addressed artifact. Both commands pass,
report zero operations and `writes=0 cli=0`, and the test asserts that no
`writes.json`, `cli_surface.json`, or `operations.json` is fabricated. The
Batch 8–10 generator's separate `http.Client.Do` regeneration path is therefore
not a missing source-import/projection capability; its absent local artifacts
remain the exact evidence gap for the 720-row exercise.

## Completed local gates

- `go test -timeout 20m ./internal/connectors/commandrunner -run '^(TestPreflightDeferredCommandReturnsNamedFoundationAfterExactTargetValidation|TestPreflightDeferredCommandFailsClosedWithoutExactTargetValidation|TestPreflightUnavailableCommandReturnsOnlyDeclaredStructuredDisposition|TestEveryImplementedCommandPassesRuntimePreflight)$' -count=1` passed.
- `go build ./cmd/pm`, `go run ./cmd/connectorgen source-import --help`, `./pm help connectors`, and `./pm connectors` passed. This internal `connectorgen` help-contract change has no `pm` manual or website surface; `docs/migration/conventions.md` is the canonical documentation update.
- `go vet ./...`, `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` passed.
- `go run ./cmd/connectorgen source-import github --out /dev/null` passed, importing 1,525 operations via the retained local reader with `writes=0 cli=0` and no tracked-file change. Its `--check` mode reports descriptor drift from the pre-existing current-main GitHub descriptor while source projection reports `writes=0 cli=0`; it was not regenerated by this scoped PR.
- `git diff --check origin/main...HEAD` passed.

## Exact-head self-review

Reviewed the final changed paths for strict JSON decoding, lock identity,
retained-artifact provenance, zero-value/slice safety, provider-I/O reachability,
and accidental execution-surface promotion. No actionable local finding. This is
not the required independent audit: request a separate Codex exact-head audit
after the PR's CI is green.

## CLI documentation parity

Not applicable to user-facing help/manual/website: this changes the internal `connectorgen` locked-source import contract. The final verification nevertheless builds `pm` and runs relevant help checks to prove no accidental public CLI regression.
