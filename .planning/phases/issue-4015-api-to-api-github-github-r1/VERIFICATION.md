# Verification — API → API GitHub route proof

Status: verified; direct PR pending.

Commands and results are appended here as they are run. Full-suite execution is
left to CI under the repository's per-command-timeout rule; all local checks
are scoped and executed separately with `-timeout 20m` where applicable.

## Planned gates

- `go test -timeout 20m ./internal/synctransport`
- `go test -timeout 20m ./internal/app`
- `go test -timeout 20m ./internal/cli`
- `go test -timeout 20m -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$' ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make agent-contract-check`
- `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, `make release-workflow-check`,
  `make connector-runtime-preflight`, and `make connector-canon-check`
- `scripts/verify-gsd-workflow 2c48e4deb34128339fccbe5d4b7daad4e13a23e7`
- fresh-binary controlled GitHub run and independent read-back

## Executed before live I/O

| Command | Result |
| --- | --- |
| `go test -timeout 20m -run '^(TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead|TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess|TestOrchestratorAdmitsEmptyResultOnlyFromExplicitSourceMarker|TestOrchestratorCommitsOnlyAfterDurableAcknowledgement)$' ./internal/synctransport` | PASS — 0.522s |
| `go test -timeout 20m -run '^TestIssueLabelTransport' ./internal/app` | PASS — 15.147s |
| `go test -timeout 20m -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$' ./internal/cli` | PASS — fresh binary lifecycle test (cached confirmation: 1.533s) |

## Live production-path proof — 2026-08-16

The fresh local binary built by `go build ./cmd/pm` was used directly; it is
the same shipped command surface, not a test harness or an HTTP substitute.
The proof project was `live-api-to-api.tmp` (git-ignored) and held the normal
encrypted PM credential only for this run. The token was captured from the
already authenticated GitHub CLI directly into an exported child-process
environment variable for `pm credentials add --from-env token=...`, then
immediately unset. It is not present in this document, the repository, status
file, or command arguments.

| Proof field | Observed value |
| --- | --- |
| Controlled retained repository | `karthik-sivadas/pm-parity-proof-api-to-api` (private) |
| Source stream / record | GitHub `issues`, run-owned issue `#1` |
| Destination action / mode / strategy | `add_issue_labels` / `full_append` / `append` |
| Destination mapping | definition-owned `target_issue -> issue_number`, `label -> singleton labels`; target issue `#2`, label `pm-api-to-api-route-r1` |
| Plan and carrier | `ETLTransportPlan` for one destructive `add_issue_labels` record; human preview token sent only through `--approval-token-stdin` to ordinary `pm etl run` |
| Happy binary result | completed `ETLRun`, `records_read=1`, `records_loaded=1`, `records_failed=0`, `batch_count=1` |
| Durable warehouse receipt | one connection-owned non-empty transport WAL, Parquet table, and receipt manifest; manifest says `records=1`, `generation=1`, with WAL/Parquet/content hashes |
| Checkpoint | persisted completed run plus the `issues` stream checkpoint with `committed_at`, after the destination acknowledgement/read-back stage |
| Independent provider read-back | `gh-axi issue list` reported issue `#2` with exactly `pm-api-to-api-route-r1`; issue `#1` had no labels. This is independent GitHub API evidence, not PM's write result. |

### Edges and refusals

- **Keyed replay:** the identical approved transport was rerun without a new
  token. It again completed one bounded record; an independent `gh-axi`
  read-back still reported exactly one `pm-api-to-api-route-r1` label — no
  duplicate or corrupt label state.
- **Explicit cleanup, not inferred delete propagation:** a separately planned,
  previewed, stdin-approved `remove_issue_label` inverse completed and
  independent GitHub read-back then reported no labels. A new cleanup plan ran
  successfully when the label was already absent, honoring the action's
  declared missing-status behavior. Replaying that cleanup approval was refused
  with `reverse plan approval has already been consumed` before another
  mutation.
- **`deletes: not_available`:** the destination declaration is retained as
  `not_available`; no source disappearance was translated into an unrequested
  delete. The only removal used the separately approved definition-bound
  inverse above. `TestIssueLabelTransportCleanupTreatsMissingLabelAsSuccessfulInverse`
  verifies that exact declared behavior.
- **Zero records:** `TestOrchestratorAdmitsEmptyResultOnlyFromExplicitSourceMarker`
  proves zero records create no stage, apply, or checkpoint effects; an
  unmarked silent zero-result source is refused. The GitHub issue-label source
  intentionally does not claim that marker.
- **Absent record mapping:**
  `TestIssueLabelTransportSourceDoesNotReadBeyondFirstPageWhenIssueMissing`
  and `TestIssueLabelTransportReadBackDoesNotReadBeyondFirstPageWhenIssueMissing`
  prove an absent configured source/target issue is refused without emitting a
  workset or reading further pages.
- **Bad closed admissions:**
  `TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess`
  asserts the specific `*SourceStreamIneligibleError` and zero
  source/stage/plan/apply/checkpoint effects. The adjacent
  `TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead` covers the
  unsupported canonical mode and verifies refusal before source I/O; the
  definition validator rejects a non-eligible apply action before executor
  registration (`DestinationTransportDescriptor.Validate`).

Focused post-live results: `internal/synctransport` PASS (0.520s),
`internal/app` PASS (18.450s), and
`internal/connectors/engine:TestWriteUnknownActionErrors` PASS (0.894s).

## Local gates

| Command | Result |
| --- | --- |
| `go build ./cmd/pm` | PASS — fresh binary SHA-256 `4d1644a590e07c0a7d23fd3067a4f3532fdaaeb49cb4f54eae0280951a75dc04`, 175611602 bytes |
| `go vet ./internal/synctransport ./internal/app ./internal/cli ./internal/connectors/engine` | PASS |
| focused `go test -count=1 -timeout 20m` transport/app/CLI/engine selectors recorded above | PASS |
| `make tidy-check` | PASS |
| `make lint` | PASS — 0 issues |
| `make docs-check` | PASS |
| `make agent-contract-check connectorgen-validate connectorgen-surface-sync connector-runtime-preflight` | PASS — 552 definitions, 0 findings; implemented-command runtime preflight PASS |
| `./connectorgen boundary . --json` | PASS — `ConnectorBoundaryReport outcome=clean`, 282 files, 552 connectors |
| `make connector-canon-check release-workflow-check github-parity-artifacts-check connectorgen-certification-matrix` | PASS — 14 GitHub artifact tests, 0 failures; certification shards current |
| `pm help etl transport`; `pm etl`; `pm etl transport github-issue-label`; `pm etl transport github-issue-label --help`; docs/website grep | PASS — existing CLI/help/docs/website carrier parity confirmed; no surface changed |
| `fm-ensure-agents-md.sh .` | PASS — `AGENTS.md` unchanged |

`go test -timeout 20m ./...` and aggregate `make verify` were deliberately
not run as one local command: repository guidance says their 550+ connector
suite routinely exceeds per-command limits, making a cutoff indistinguishable
from a hang. The focused packages plus every non-test `make verify` gate above
ran locally; CI remains responsible for the full suite.

## Historical repository decision — resolved

Blocked before any PM credential injection or provider write. The authoritative
runbook says the recorded fine-grained token is revoked (`HTTP 401`) and must
not be re-tested. The live GitHub App installation is reported working, but no
controlled repository is available to it. A narrowly named run-owned repository
creation attempt through `gh-axi` was refused by GitHub with
`karthik-sivadas does not have the correct permissions to execute
CreateRepository`; no repository, issue, label, or PM project was created.

That failed organization-repository creation attempt left no provider residue.
Captain instead authorized the retained private personal repository below; the
GitHub connector's supported token authentication makes a GitHub App
installation unnecessary for this narrow route proof.

## Authentication decision — personal proof repository

Captain authorized and `gh-axi` created the private, retained evidence
repository `karthik-sivadas/pm-parity-proof-api-to-api`; its run-owned source
and destination sentinel issues are `#1` and `#2`. The proof label was applied,
independently read back, replayed safely, then removed through its explicit
approved inverse; the repository and sentinel issues remain retained.

The GitHub connector supports token bearer authentication and GitHub App
installation authentication as independent production paths. The bundle's
first auth candidate is `secrets.token`; it wins whenever a token is supplied
(`internal/connectors/defs/github/streams.json:9-18`), and the token is
declared as an accepted GitHub credential secret
(`internal/connectors/defs/github/spec.json:66-69`). The closed transport does
not constrain authentication: it resolves the destination's normal credential
runtime and invokes the generic writer
(`internal/app/issue_label_transport_approval.go:435-478`, `:331-337`). Both
auth choices run that same engine runtime and declarative write/read path;
they differ only in how the bearer credential is obtained
(`internal/connectors/engine/read.go:543-578`,
`internal/connectors/engine/auth.go:84-117`,
`internal/connectors/hooks/github/hooks.go:66-117`).

Accordingly, this route proof uses the already-authenticated `karthik-sivadas`
GitHub CLI identity through the connector's supported `token` credential
path. It does not use the revoked runbook token, and does not claim GitHub App
or installation-only capability certification.

## Resolved credential handoff

The authenticated GitHub CLI reports the expected `karthik-sivadas` identity
and `repo` scope. The initial PM credential handoff correctly refused before
credential creation because the shell variable containing the CLI-managed token
was not exported to the `pm` child process (`environment variable
PM_API_TO_API_GITHUB_TOKEN is empty`). No PM credential, connection, provider
label write, or checkpoint was created by those attempts. This was resolved by
capturing the existing CLI token through command substitution directly into an
exported child-process environment variable, then unsetting it immediately
after `pm credentials add`; the secret is never rendered, logged, persisted in
evidence, or passed via argv.
