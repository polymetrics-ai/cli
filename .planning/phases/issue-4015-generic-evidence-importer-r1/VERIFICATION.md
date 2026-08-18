# Verification Checklist — Generic Certification Evidence Importer

- [x] Certified diagnosis rechecked against current integration-base code.
- [x] Focused `./cmd/connectorgen` tests exercise redaction, generic import, missing evidence, broken evidence, and definition-derived second shape.
- [x] Existing PostgreSQL accepted evidence is byte-identical: all 14 PostgreSQL records have no diff from rebased integration head `c9791db4d`.
- [x] GitHub matrix remains at zero published capabilities until a rotated credential produces a new report plus matching redacted proof. The generic schema-v2 `observed_operations` path is staged, but it cannot honestly manufacture the missing artifact for the 34 prior reads.
- [x] Deliberate broken-evidence command result captured, followed by restoration: PostgreSQL CDC went red (`live_tested=false, records=0`) after its record key was broken, then green (`true, 1`) after restoration.
- [x] Secret-scanner negative control: a synthetic secret-shaped certification string made `connectorgen validate` fail with `[secret_literal]`; it was removed and the same command returned zero findings.
- [x] Red: CI job `95424260739` failed at `make certify-timing` because its declaration-stage-selection test returned unselected candidates, not because the harness exceeded its budget (`16` real CLI invocations, budget `25`).
- [x] Green: `make certify-timing` passed after the selector correction in 21 seconds, with the same `16` real CLI invocations (budget `25`) and the unchanged 3m30s duration cap.
- [x] Focused corrected-tree checks passed: `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/certify`, `go run ./cmd/connectorgen certification-candidates --connector github --check`, `go run ./cmd/connectorgen certification-matrix --check`, and `git diff --exit-code c9791db4d10073099006fe12f01723adc6bac098 -- internal/connectors/certifications/evidence`.
- [x] `make verify` passed after the selector correction: formatting, tidy, vet, full `go test -timeout 20m ./...`, build, docs, smoke, lint, agent contract, generators, certification, boundary, canon, and release-workflow gates. Live GitHub checks remain deliberately deferred pending rotation.
- [x] Targeted schema-v2 checks passed: `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/certify ./internal/cli`, plus the final consumer-package run `go test -timeout 20m ./cmd/connectorgen -count=1`.
- [x] Generated documentation was regenerated twice and byte-stable: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`, plus `pnpm run gen:docs` and `pnpm run gen:website-data` from `website/`.
- [x] GSD inline `verify-work` and `code-review` fallback completed: no actionable implementation, accounting, path-safety, or secret-redaction finding remains. The external review route is `claude_auto` after the direct PR opens.
- [x] PR API base read-back equals `integration/4015-mvp-flat-r1`: `gh-axi pr list -R polymetrics-ai/cli --state open --base integration/4015-mvp-flat-r1 --head fm/cli-generic-evidence-importer-r1` returned only PR #4216; `gh-axi pr view 4216 --full` returned the complete issue-linked body.
- [x] Red: newest Verify job `95429332235` identified four stale golden transcript entries, not a harness-cost regression. The package completed in 1082 seconds, below the known 1200-second ceiling.
- [x] Green: supported targeted regeneration updated exactly those four entries. The second generator pass was byte-stable at SHA-256 `ffe235d7ae35274de4de5d61bbf3d01301e25106efbd5d9349a938f76dc79045`, and the selected normal test passed. The complete local transcript corpus timed out after its unchanged 20-minute cap, so the PR's newest Verify run remains the complete-corpus authority.
- [ ] Pending CI: newest Verify run after the golden fixture update must complete successfully. No local timeout or test bound was changed.

## Credential-rotation limitation

- The prior GitHub classic PAT was exposed to local terminal output from an
  external runbook and is revoked / do-not-use. No live call will use it. The
  34 prior direct-read passes have no recoverable report-plus-proof artifact,
  so the staged binding cannot yet publish evidence. A credential owner must
  rotate it before a fresh bounded proof run. This does not block credential-free
  compilation, generation, or local validation.

- `security/snyk` is not a local Make target in this checkout. Its failure is
  the known identical base-branch CI condition supplied by the delivery brief;
  this change neither invokes nor masks it.
