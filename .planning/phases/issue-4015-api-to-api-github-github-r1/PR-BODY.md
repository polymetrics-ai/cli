Refs #4015

## Intent

Prove the narrow API → API GitHub route through the shipped `pm` binary:
GitHub `issues` → durable warehouse → GitHub issue labels. This proves the
route only. GitHub write coverage remains **two actions of 607**.

## What Changed

- Added GSD/TDD, UAT, review, and sanitized live-proof evidence for the
  definition-owned GitHub `issue_label_destination` route.
- No production connector, PostgreSQL transport, #4184 atomicity, generated
  surface, generic writer, or GitHub action changed.

## Production-path proof

- Controlled retained private repository:
  `karthik-sivadas/pm-parity-proof-api-to-api`.
- Authentication: the connector's supported GitHub `token` bearer production
  path, supplied only through PM's environment-backed credential input. No
  credential or approval value is in this PR, evidence, argv, or status file.
- Source: GitHub `issues`, run-owned issue `#1`.
- Destination: `issue_label_destination`; `full_append` → `append` →
  `add_issue_labels`; definition-owned mapping `target_issue → issue_number`,
  `label → singleton labels`; target issue `#2`.
- Happy result: fresh built `pm` completed one read/loaded record through the
  WAL, Parquet receipt/reopen, acknowledged destination, and checkpoint.
  Independent `gh-axi` read-back observed exactly
  `pm-api-to-api-route-r1` on issue `#2` (and no label on source issue `#1`).

## Bad and edge coverage

- Bad — ineligible stream returns the specific typed
  `*synctransport.SourceStreamIneligibleError` before any source, stage, plan,
  apply, or checkpoint I/O. Unsupported canonical mode and a destination
  apply action outside the positive allowlist are refused by preflight/
  descriptor validation before executor I/O.
- Edge — zero records are accepted only from an explicit source marker and
  otherwise have no stage/apply/checkpoint side effect; the GitHub source does
  not silently represent a missing configured issue as an empty result.
- Edge — a configured source or target issue absent from the first provider
  page emits no workset and performs no extra-page search.
- Edge — replaying the exact approved live transfer retained exactly one
  independently read GitHub label: no duplication or corruption.
- Edge — destination `deletes: not_available` was honored: no implicit delete
  propagation exists. The separately planned/approved `remove_issue_label`
  inverse removed the label; a missing-label inverse succeeded under its
  declared missing status; replaying its consumed approval was refused before
  another mutation.

## GSD / TDD / skills

- Inline/manual GSD fallback was used because the canonical single-worker
  contract forbids compatible role spawning. Resolved and executed through
  `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` for
  `issue-4015-api-to-api-github-github-r1`.
- Red: existing fail-closed regressions establish typed ineligible-stream,
  unsupported-mode, zero-result, and no-I/O behavior before live provider I/O.
  Green: fresh-binary live route plus independent GitHub read-back. Refactor:
  replay and explicit cleanup/missing-label/replayed-approval edges.
- Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, and `golang-lint`.

## CLI/help/docs parity

No CLI surface changed. Verified `pm help etl transport`, bare `pm etl`, bare
`pm etl transport github-issue-label`, its `--help`, and existing CLI/website
route documentation references.

## Testing

- `go build ./cmd/pm`
- focused `go test -count=1 -timeout 20m` for `internal/synctransport`,
  `internal/app`, `internal/cli:TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle`,
  and `internal/connectors/engine:TestWriteUnknownActionErrors`
- `go vet ./internal/synctransport ./internal/app ./internal/cli ./internal/connectors/engine`
- `make tidy-check lint docs-check agent-contract-check connectorgen-validate connectorgen-surface-sync connector-runtime-preflight connector-canon-check release-workflow-check github-parity-artifacts-check connectorgen-certification-matrix`
- `./connectorgen boundary . --json` — clean, 0 findings

The aggregate `go test -timeout 20m ./...` / `make verify` was not invoked as
one local command because repository guidance says its 550+ connector suite
exceeds the per-command limit; focused tests and each non-test gate ran above.

## Delivery and review

- Base: `integration/4015-mvp-flat-r1` at `2c48e4deb`.
- Direct PR; no no-mistakes run.
- Automated review route: `claude_auto` pending this trusted-author,
  non-draft PR opening. No manual Claude/Copilot request was posted.
