# Verification checklist — issue 4347 source-lock usability

- [x] Focused red tests recorded for missing form pin, canonical reserialisation, 403, and undersize.
- [x] Focused red/green test recorded for the distinct 403/TLS `BOT-BLOCK` verdict.
- [x] CI failure verdict: clean `060bb7864` passes `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceImportRetainedArtifactRejectsMissingAndMismatchedCopies$'`; the branch-owned root fix passes that test plus `TestSourceRetainReportsBotBlockBeforeWrongSourceOrDrift`.
- [x] `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceRetain'` passes after implementation. The broader `go test -timeout 20m ./cmd/connectorgen` was stopped locally after its unrelated `TestCertificationMatrix...` fixture started `certification-matrix --all` with a fresh `GOCACHE`; it produced no source-lock failure. CI carries that broader package path.
- [x] `go vet ./cmd/connectorgen` passes.
- [x] Built `connectorgen` and ran `source-retain` sequentially for fastly, github, hubspot, pipedrive, shipstation, squarespace, woocommerce, and zendesk-support; all commands exit 0 and pre/post SHA-256 values prove their locks were unchanged.
- [x] `go run ./cmd/connectorgen surface-sync --check` passes.
- [x] Relevant independent `make verify` gates pass: `tidy-check`, `lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`, `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, and `release-workflow-check`. Do not run aggregate `go test ./...`.
- [x] `scripts/gsd prompt verify-work 4347` and `scripts/gsd prompt code-review 4347` ran as the recorded inline/manual fallback; `REVIEW.md` records the review disposition.
- [ ] PR base is read back through the GitHub API and exactly equals `main` after the follow-up push.

## Independent audit R1 gap closure — pending

- [x] F1 red and green: strict v3 document-owned operation-evidence identity preserves six lanes and declared/deferred rows. Exact focused green command passed in 2.822s on 2026-08-26.
- [x] F2 red and green: generic rendered publication citation rejects unless fragment or verified capture extraction binding is present. Read-only Batch 6–7 impact at `origin/fm/cli-map-batch67-r1` / `18248d233e6abd9d7ec03075a225cf35ee2f5399`: 861 generic citation rows in eight connectors are intentionally not admitted until their lock owners add a fragment or binding.
- [x] F3 red and green: HTTP MIME/body evidence rejects plausible login and `Error 503` pages plus invalid/bad MIME as wrong-source before drift, without rejecting legitimate documentation HTML.
- [x] `go test -timeout 20m ./cmd/connectorgen` passes after the final repair (158.269s on 2026-08-26).
- [x] `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `make tidy-check`, and `make docs-check-no-build` pass after the repair.
- [x] Clean tracked archive at `9e1bfdb9b21ab346f84537bfb094a22782b0d5d5` passed `agentcontractgen check`, `connectorgen validate`, `surface-sync --check`, `operation-evidence --check` (1,525 rows; fixed-100 passed), certification subject/matrix/candidates/sweep checks, and `connectorgen boundary . --json`. Its temporary archive was deleted after the checks; it excluded the preserved live-retention artifacts.
- [x] `make lint` passes after the static-analysis repair. No aggregate `go test ./...` was run.
- [ ] PR #4350 is pushed, its API-reported base is `main`, and Firstmate is asked for a fresh independent audit. No merge is performed.

## Independent exact-SHA audit R2

- [x] Red and green: populated v3 source documents cannot be classified as absence or suppress REST evidence.
- [x] Red and green: canonical JSON served as HTML is wrong-source before canonical identity drift.
- [x] Red and green: rendered citation fragments match the operation location and supplied bindings are never ignored.
- [x] Red and green: canonical JSON rejects duplicate object members as ambiguous bad-source input, not ordinary drift.
- [x] Red and green: canonical JSON accepts formatted/minified equivalent JSON
  with unequal raw byte counts; raw identity is manifest provenance only while
  byte-identity locks remain exact.
- [x] Focused suites pass: `go test -timeout 10m ./cmd/connectorgen -run '^TestSourceRetain' -count=1` (1.409s) and the operation-evidence/rendered-reference/canonical suite (17.618s).
- [x] Full changed suite passes: `go test -timeout 20m ./cmd/connectorgen -count=1` in the clean tracked-only archive (279.084s). Aggregate `go test ./...` was not run.
- [x] `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `go build -o .task-bin-connectorgen ./cmd/connectorgen`, `make tidy-check`, `make docs-check-no-build`, `make smoke-no-build`, and `make lint` pass. `git diff --check` passes.
- [x] Clean tracked-only archive check passes: `agentcontractgen check`, source validate, surface sync, operation evidence (1,525 rows; fixed-100), certification subject/matrix/candidates/sweep, boundary (553 connectors; 317 files; no findings), connector canon, GitHub parity artifacts, and all release workflow scripts including installed GitHub certification.
- [x] Root `agentcontractgen check` is not a code failure: it correctly refuses the deliberately preserved untracked `.fm-main-clean.qazhOS/.claude/agents/pm-connector-worker.md` duplicate. The exact command passes in the clean archive containing this branch's diff and no untracked artifacts.
- [x] PR #4350 is open against `main` with head branch
  `fm/cli-source-lock-usability-foundation-r1`; `gh-axi pr list --state open
  --base main --head fm/cli-source-lock-usability-foundation-r1 --limit 10`
  returned only #4350. Before the delivery-record commit, `git ls-remote
  origin refs/heads/fm/cli-source-lock-usability-foundation-r1` and local HEAD
  both returned `54c2e653a088e809884ada50ab41a9e54c3f90b5`. A PR comment requests
  a fresh independent audit; no merge was performed. The body editor is
  blocked by GitHub's deprecated `repository.pullRequest.projectCards` GraphQL
  field, so that immutable PR comment is the R2 delivery record.

## Independent exact-SHA audit R3 help-contract repair

- [x] Red first: at exact remote head `f8f6240792fa4364c4890b6ad3fc30bff1a33db6`, `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceRetainHelpAndMigrationDocumentationDescribeIdentityAndWrongSource$' -count=1` failed because rendered `source-retain --help` split `wrong source` across a newline.
- [x] Green: that exact public-help/migration-document contract passed in 1.123s after a usage-text-only reflow.
- [x] Full changed suite: `go test -timeout 20m ./cmd/connectorgen -count=1` passed in 156.924s. Aggregate `go test ./...` was not run.
- [x] Static/generation checks passed: `go vet ./cmd/connectorgen`; tracked-only `connectorgen validate`, `surface-sync --check`, `operation-evidence --check` (1,525 rows; fixed-100), and `certification-subject --check`; `make docs-check-no-build`; and `git diff --check`.
- [x] The preserved untracked retained-artifact evidence causes root `certification-subject --check` to report local generated-state drift. No artifact was regenerated or staged; the same check passed from a tracked-only archive containing this repair.
- [ ] Commit/push the minimal repair, record its exact remote head and CI start, and request a fresh independent exact-head audit. No merge.

## Independent exact-SHA audit R4 duplicate-member repair

- [x] Red: both duplicate-member orders prove source-retain could select a URL or identity before rejection, and could call its fetch seam.
- [x] Red: populated-then-empty duplicate `rest.source_documents` proved operation evidence could project a dynamic schema-v3 lock as absence before rejection.
- [x] Green: retain and operation-evidence reject recursively duplicated lock JSON through the existing tolerant duplicate validator; source import remains strict. Focused command passed in 1.111s.
- [x] Focused source-retain and operation-evidence tests pass: `^TestSourceRetain` (1.298s) and `^TestOperationEvidence` (16.974s). The completion-tracked `go test -timeout 20m ./cmd/connectorgen -count=1` passed in 173.360s with exit 0; its command output and exit marker are in the Firstmate task status file.
- [x] `go vet ./cmd/connectorgen` and `git diff --check` pass. Generator checks pass: `source-import github --check` (1,525 operations, 0 inbound events), `operation-evidence --check` (1,525 rows; 5 rollups; fixed-100), and `surface-sync --check` (553 connectors; zero drift). `certification-subject --check` is stale only in this root because preserved untracked retain artifacts participate in its input; it and `agentcontractgen check` pass read-only from the preserved tracked-only archive.
- [x] `go build ./cmd/pm`, `make tidy-check`, `make docs-check-no-build`, and `make smoke-no-build` pass. `make lint` was retried after the shared linter lock cleared and passed with 0 issues; `go run ./cmd/connectorgen validate` checked 553 connectors with 0 findings.
- [x] Inline/manual GSD `verify-work` and `code-review` prompts were resolved after verification. `REVIEW.md` records the R4 scope, security boundary, and no-actionable-finding disposition.
- [x] Committed/pushed R4 parser/test/evidence repair `99ef2eb9224cefcfb48f0721731abf115f1e276f`. `gh-axi pr list --state open --base main --head fm/cli-source-lock-usability-foundation-r1 --limit 10` returned only #4350, proving the API-backed PR remains open against `main`.
- [x] The `gh-axi pr edit` body route is blocked by GitHub's deprecated `repository.pullRequest.projectCards` GraphQL field. The accepted fallback is a PR #4350 comment carrying the exact R4 SHA, red/green outcome, completion-tracked full-suite result, and fresh independent-audit request. No merge.
