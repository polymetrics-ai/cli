## Intent

Generator capability change for the connector artifact materializer. Every
operation in a valid OpenAPI 3.x or Swagger 2.0 artifact is represented; a
missing executor becomes a visible `not_implemented` command with a
machine-checkable `named_dependency=<slug>` note; and source-surface rows that
the artifact omits remain present with
`present-in-surface-absent-from-artifact`.

Closes #3958. The GSD phase evidence in
`.planning/phases/persistiq-artifact-materialize-pilot-r1/` records the
delivery contract and red/green proof.

## What Changed

- Materialization retains complete artifact operation inventory instead of
  refusing a connector when a source surface is broader than its published
  artifact.
- Command availability is honest: only runtime-preflightable commands are
  `implemented`; unsupported commands remain visible with named dependencies.
- Static validation checks the named-dependency contract for every
  `not_implemented` command.
- Surface schema/model supports the exact discrepancy marker.
- Materialization is cheap; the repository-wide runtime preflight belongs to
  the single final `batch gate`, not each candidate.
- Added red/green tests and strict artifact-version parsing.
- PersistIQ pilot evidence is included. The 392-connector generation is a
  separate follow-up and is intentionally not in this PR.

## Testing

- `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner` — pass.
- `go vet ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner` — pass.
- `go build ./cmd/pm` — pass.
- `go run ./cmd/agentcontractgen check` — pass.
- `connectorgen validate internal/connectors/defs --json` — 551 connectors, 0 findings, 0 warnings.
- `connectorgen surface-sync --check` — 551 scanned, 0 fields filled/corrected.
- `TestEveryImplementedCommandPassesRuntimePreflight` — pass.
- PersistIQ: 21 mapped, 21 implemented, 3 named-dependency, 3 flagged-discrepancy, 24/24 real-binary command paths reachable, 0 failed.

PersistIQ rerun wall clock: identify 0.03s; map 0.03s; fetch/digest/parse
2.70s; materialize + static/runtime/binary gates 50.07s; report 0.09s;
total 52.92s. Artifact: OpenAPI 3.0.1, 47,796 bytes, SHA-256
`0bf3e1ecbfbf6215360b5bb8f9d4fda816df4e1872470a00b529fb3e8b80946f`.

## Safety and certification

No credentials were read, requested, printed, or stored. No provider operation
was exercised. **Implemented, not certified, never exercised against the
provider.** Certification is withheld for every connector; this PR certifies
neither PersistIQ nor any existing connector surface.

## Delivery record

- GSD lifecycle evidence: `DISCUSSION-LOG.md`, `PLAN.md`, `TDD-LEDGER.md`,
  `VERIFICATION.md`, `SUMMARY.md`, and `PR-EVIDENCE.md`.
- Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation`.
- Inline/manual GSD fallback was recorded because the task runtime does not
  provide a compatible isolated GSD worker; red/green evidence and all gates
  were still run.
- Commits were pushed to `fm/cli-mass-artifact-materialize-r1`; no default
  branch push or merge was performed.
- Automated review route: `claude_auto` on PR open; no Copilot fallback has
  been requested.

## Follow-up

After this generator capability lands and is reviewed, fetch/materialize the
eligible 392 in review-sized batches and run one combined final gate. Do not
interpret this PR as the 392 generation or as provider certification.
