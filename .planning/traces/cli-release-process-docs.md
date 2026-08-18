# Release Process Documentation Delivery Trace

## Task Delivery Header

- Issue: Refs #4015 — Production MVP
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with repository documentation and release checks green.
- Working branch: `fm/cli-release-process-docs`
- Task: Document the maintainer release process, register it in the repository-defined documentation surface, correct the stale `.goreleaser.yaml` workflow comment, audit other release references, and make no release-behavior or version changes.
- Verification: Check the documentation against `.github/workflows/release.yml`, `release-please-config.json`, `.release-please-manifest.json`, `scripts/release-please-pm-filter.py`, and release scripts; run the repository documentation, link, release-workflow, generated-file, snapshot, and relevant test gates; verify the opened PR base through the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Release process is documented in the audience-classified location and registered | live | Repository structure classifies the procedure as maintainer documentation, `CONTRIBUTING.md` registers it, and direct link plus workflow/config assertions confirm the documented trigger, versioning, jobs, targets, artifacts, checks, and recovery procedure. |
| The stale workflow comment names a real source of truth | live | Repository search finds no `.goreleaser.yaml`, and the corrected comment points to checked-in release assembly/build code. |
| Documentation and release gates pass | live | The trace records exact commands and successful results for the repository-defined relevant gates. |
| The PR targets the required integration branch | live | GitHub API reports `integration/4015-mvp-flat-r1` as the opened PR base. |

## Documentation method

- Required skill: `golang-documentation`.
- GSD lifecycle: manual documentation-only path; no runtime behavior changes, so the behavior-changing `discuss-phase → plan-phase --tdd → execute-phase → verify-work → code-review` sequence is not applicable. This trace records plan, source verification, validation, and review evidence.
- Assertion rule: checks must inspect the changed documentation/comment or exercise the release/documentation gates; command success without the claimed observable is not accepted as factual proof.

## Verification record

### Source verification

- `git ls-tree -r --name-only origin/main` and the same command for
  `origin/integration/4015-mvp-flat-r1` found no `.goreleaser.yaml` or
  `.goreleaser.yml`.
- Direct reads of `.github/workflows/release.yml`,
  `release-please-config.json`, `.release-please-manifest.json`,
  `scripts/release-please-pm-filter.py`, `scripts/assemble-release-assets.sh`,
  `scripts/verify-release-assets.sh`, and `docs/release-verification.md` supplied
  every repository-specific factual claim in `docs/releasing.md`.
- A Python assertion over the release config confirmed Go release type, package
  `pm`, title pattern, `bump-minor-pre-major: true`, and manifest `0.1.1`.
  Release Please's official manifest documentation confirmed that `feat`
  normally requests a minor bump and `bump-minor-pre-major` maps pre-1.0
  breaking changes to minor; together these prove the documented
  `0.1.1 + feat → 0.2.0` example.
- A Python assertion over the workflow confirmed all event/job names, all four
  target/runner triples, release flags, and one-day intermediate artifact
  retention.
- Repository search found no dedicated maintainer-doc audience registry or
  Markdown link checker. The existing classification is contributor/maintainer
  release policy in `CONTRIBUTING.md`, end-user usage in `docs/GUIDE.md` and the
  website, and release trust operations in `docs/release-verification.md`.
  Therefore the maintainer runbook is registered from `CONTRIBUTING.md`; a
  direct local-Markdown-link assertion checked all changed Markdown links.

### Commands and results

- `make fmt` — pass; no Go source changes.
- `make tidy-check` — pass; `go.mod` and `go.sum` unchanged.
- `make docs-check` — pass; binary built and 552 connector docs validated.
- `make release-workflow-check` — pass; pinned dependencies, release
  notification/filter assertions, and four-target parity passed.
- `go run ./cmd/agentcontractgen check` — pass; canonical contract and generated
  projections current.
- `go run ./cmd/connectorgen surface-sync --check` — pass; 552 connectors
  scanned with zero corrections.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass;
  552 connectors, no findings or warnings.
- `go vet ./...` — pass.
- `make lint` — pass; zero issues.
- `make smoke-no-build` — pass; complete local extract/query/write smoke path.
- `go test -timeout 20m ./...` — pass twice; the final run followed a rebase onto
  integration commit `e06cdfdf7` and included `internal/cli` (643.300 seconds).
- `make github-parity-artifacts-check` — pass; tests and generated GitHub parity
  artifacts current.
- `make connectorgen-certification-matrix` — pass.
- `make connectorgen-certification-candidates` — pass.
- `make connectorgen-certification-sweep` — pass; 1,571 GitHub commands current.
- `make connector-boundary` — pass.
- `make connector-canon-check` — pass.
- `git diff --check` — pass.
- Local changed-Markdown link assertion — pass.
- After `origin/integration/4015-mvp-flat-r1` advanced,
  `git rebase origin/integration/4015-mvp-flat-r1` completed cleanly. `make docs-check`,
  `make release-workflow-check`, all GitHub parity/certification generated-file
  checks, surface sync, definition validation, and the full Go suite passed
  again on the rebased head.

### Stale release-reference audit

Changed:

- `.github/workflows/release.yml` pointed at the nonexistent
  `.goreleaser.yaml`; it now names the native workflow build as the linker-flag
  source of truth and states that assembly does not rebuild.
- `packaging/linux/nfpm.yaml` claimed to mirror an absent GoReleaser `nfpms`
  block; it now names the nFPM file and checked-in assembler/verifier/tests as
  the package contract.
- `docs/release-verification.md` described current package assembly as a
  GoReleaser/nFPM build; it now describes the actual archive and nFPM assembly.

Left unchanged:

- `scripts/assemble-release-assets.sh` and the opening comments in
  `packaging/linux/nfpm.yaml` compare the current approach with GoReleaser OSS
  and explain why GoReleaser is not used. Those references are accurate.
- `website/lib/blog.ts` says GoReleaser builds artifacts inside an article whose
  evidence snapshot is explicitly dated 2026-07-16. It is historical editorial
  content, not current release instructions, so rewriting it as present-day
  machinery would falsify the dated narrative and expand this maintainer-doc
  task into website content regeneration.
