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
- [ ] Exactly 720 source-cited deferred rows reconcile as 187 Amplitude + 49 Dremio + 193 Ashby + 84 Workable + 207 HiBob. Blocked: the current Batch 8–10 branch has only locks/crosswalks/dispositions, no `sources/artifacts/` files or retained-artifact manifests for these five source locks.
- [ ] Rows preserve exact source citation, provider operation ID, stable identity, applicable lane, and exactly one `missing_foundation` disposition.
- [ ] Deferred command preflight stops before credential/transport/record/mutation; retained runnable command boundary still reaches missing credential.
- [ ] Generator/projection/evidence/surface-sync/validate and bounded JSON duplicate invariants pass.
- [ ] Formatting, vet, build, diff, individual repository gates, rebase, exact-head review, and PR API base verification pass.

## Foundation evidence

- RED: `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportV3RenderedReferenceCoverageOnlyRequiresExplicitClosedProof$' -count=1` failed before implementation with `json: unknown field "coverage_only"`.
- GREEN: `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportV3RenderedReferenceCoverageOnly' -count=1` passed.
- Regression: `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportVersion3RenderedReference|TestSourceRetainRetainsRenderedReferenceAndBundleArtifacts|TestSourceImport_RejectsUnknownSectionAndIndependentIndexOverflow)$' -count=1` passed.
- Full-package correction: `go test -timeout 20m ./cmd/connectorgen -count=1` first failed only because the existing `TestSourceImportVersion3RenderedReferenceRetainsCapturedEvidenceWithoutOperations` fixture omitted the newly required discriminator. The fixture now declares `coverage_only:true`; its focused regression with all coverage-only tests passes. The corrected full package passed in 193.136 seconds.
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

## CLI documentation parity

Not applicable to user-facing help/manual/website: this changes the internal `connectorgen` locked-source import contract. The final verification nevertheless builds `pm` and runs relevant help checks to prove no accidental public CLI regression.
