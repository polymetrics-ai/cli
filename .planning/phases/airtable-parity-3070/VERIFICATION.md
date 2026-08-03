# Verification checklist — Airtable official API parity

## Prior required gates

These gates passed before the review-hardening changes and are retained as historical evidence; the outer pipeline owns authoritative full-tree validation for the final tree.

- [!] `go run ./cmd/connectorgen validate internal/connectors/defs/airtable` — current `connectorgen validate` treats the argument as a directory of bundle directories, so this exact single-bundle path validates `fixtures/` and `schemas/` as fake connectors and exits 1 with missing `metadata.json`. No tooling/runtime behavior was changed for this local-only connector wave.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — `connectorgen validate: 549 connector(s) checked, 0 findings`.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/airtable' -count=1` — `ok polymetrics.ai/internal/connectors/conformance`.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — `ok polymetrics.ai/internal/cli`.
- [x] `go build ./cmd/pm`.
- [x] `make connector-boundary` — outcome `clean`.
- [x] `make verify` — passed on cached retry after one transient full-suite timeout in `internal/connectors/certify` under parallel `go test -timeout 20m ./...`; package passed standalone and final `make verify` completed successfully.
- [x] `git diff --check`.

## Fixture-only safety

- No live provider calls.
- No credentials requested or used.
- No Airtable writes executed.
- Certification metadata is fixture/candidate-only; no live certification claim was made.
- No push, PR, `/no-mistakes`, VPS, Thaalam, or provider-side operations were run.

## Results

Implemented fixture-only Airtable parity partition: 103 official OpenAPI operations tracked as 28 stream-backed GET/read/changefeed operations, 44 typed write actions, 1 HyperDB direct-read operation/CLI command, and 30 blocked operations. Comments use the exact official per-record endpoint, webhook replay uses the executable `limit=50`, and attachment upload remains blocked on `airtable-bounded-base64-upload-foundation` instead of exposing an unbounded write.

## Review-hardening gate

- [x] `go test ./internal/connectors/defs ./internal/connectors/conformance -run 'Airtable|Conformance/airtable' -count=1` — passed for both focused packages.

## CI issue-link guard repair — 2026-08-02

- [x] GSD adapter preflight: `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase airtable-parity-3070 --dry-run` remained unavailable with `unknown GSD command: programming-loop`, so this slice reused the phase's recorded manual-GSD/TDD fallback.
- [x] Red: the focused checkpoint regression test failed with the reported `PR body must reference an issue` violation before production changes.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` — passed.
- [x] `go vet ./internal/coordination/issueguard ./cmd/prissueguard` — passed.
- [x] Exact PR #3540 title/body through `go run ./cmd/prissueguard` — `issueguard: ok (8 linked issues)`.
- [x] Negative coverage keeps standalone GitHub issue URLs, incomplete checkpoint wording, and vague `Issue`/`References` relationships rejected; positive coverage passes with LF and CRLF bodies.
- [x] `git diff --check` — passed.

The outer no-mistakes executor still owns commit, push, PR-body mutation, broader validation phases, and authoritative hosted CI rerun.

## CI connector-boundary repair — 2026-08-02

- [x] Red: `make connector-boundary` reported one `connector_literal` finding for `github` in the canonical issue-host regex at `internal/coordination/issueguard/guard.go:34`.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` — passed, preserving the exact PR #3540 checkpoint acceptance and negative cases.
- [x] `make connector-boundary` — `outcome: clean`, 130 shared files and 550 connectors checked with zero findings or warnings.
- [x] `make verify` — passed end-to-end: format, tidy check, vet, full tests, build, docs validation, smoke, lint, connector validation, connector boundary, and release notification assertions.
- [x] No live Airtable calls, provider credentials, provider writes, dependency changes, PR mutation, push, or pipeline-control commands were used.

## CI checkpoint-indentation repair — 2026-08-02

- [x] Live read-only PR inspection confirmed the hosted body indents the exact canonical-section heading and first issue URL by four spaces.
- [x] Red: `go test ./internal/coordination/issueguard -run TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks/GitHub-indented_body -count=1` failed with the reported missing-issue violation before the production regex changed.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` — passed with standard and GitHub-indented LF/CRLF fixtures.
- [x] `go vet ./internal/coordination/issueguard ./cmd/prissueguard` — passed.
- [x] Exact PR #3540 title/body through `go run ./cmd/prissueguard` — `issueguard: ok (8 linked issues)`.
- [x] `make connector-boundary` — passed with a clean report.
- [x] `make verify` — passed end-to-end on the final tree.
- [x] No live Airtable calls, credentials, provider writes, dependencies, PR mutation, push, or no-mistakes pipeline-control commands were used.

## CI phase repair — unvalidated-checkpoint heading indentation — 2026-08-03

Notably, the prior indentation fix only taught the canonical-section heading and URL patterns to
accept 0–4 leading spaces; the blocking `## Unvalidated cloud checkpoint — do not merge yet` heading
pattern still required zero leading whitespace, so the hosted PR resumed failing
`require-linked-issue` once that heading was indented.

- [x] Red: reverting the guard and running `TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks` reproduced the missing-issue violation on the new `indented checkpoint heading` and CRLF fixtures.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1` — passed, including new indented-heading LF/CRLF fixtures that pin the exact failing shape.
- [x] `go vet ./internal/coordination/issueguard ./cmd/prissueguard` — passed.
- [x] `go build ./cmd/prissueguard` — passed.
- [x] Branch-name gate: verified `fm-airtable-import-tmp` and `fm/cli-...` pass the `conventions.yml` branch-name case while a genuinely malformed branch is still rejected.
- [x] `git diff --check` — passed.
- [x] No live Airtable calls, credentials, provider writes, dependencies, PR mutation, push, or no-mistakes pipeline-control commands were used by this CI phase.
