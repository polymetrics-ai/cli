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

## Generalization validation before merge

Three eligible, deliberately different shapes were staged as validation
evidence only; no generated production connector surface was added:

| Connector | Shape | Mapped | Implemented | Named dependency | Discrepancy | Reachable | Result |
|---|---|---:|---:|---:|---:|---:|---|
| watchmode | 23-read OpenAPI 3.0.3 | 23 | 13 | 32 | 22 | 45/45 | pass |
| docuseal | 7-read/16-write OpenAPI 3.1.0 | 0 | 0 | 0 | 0 | 0 | **failed: top-level webhooks rejected as artifact inventory unknown** |
| float | 44-read/51-write Swagger 2.0 | 0 | 0 | 0 | 0 | 0 | **failed: external path-item reference not exhaustively resolvable** |

Watchmode mapped all 23 artifact operations as `direct_read` with 0
unclassified, retained 22 source-surface-only rows using the exact
`present-in-surface-absent-from-artifact` marker, and did not refuse the
connector. Its static gates were clean and the real binary reached all 45
command help paths (13 implemented and 32 visible not-implemented commands).

Watchmode wall clock: identify 0.04s; fetch/digest 2.52s; map 0.02s; batch
plan 1.78s; materialize/parse 0.65s; validate 0.67s; surface-sync
derive/check 0.68s/0.64s; batch gate 0.66s; runtime-preflight regression
5.28s; binary build 9.71s; bare namespace 2.48s; 45-command reachability
54.70s; report 0.06s; total 79.89s.

DocuSeal fetched as OpenAPI 3.1.0 (192,929 bytes; SHA-256
`7ac10d1c39b335bce962b6de277d88aded8ce476518b83835c76ad80157e0e4b`) but
materialization rejected its 11 top-level webhooks before mapping. Float
fetched as Swagger 2.0 (8,634 bytes; SHA-256
`d204eae066136386aea4ea955fb9d0d08ef9ca85eafabc2bb2dcd30b8751211c`) but
materialization rejected an external path-item reference before mapping.

**Generalization result: NOT READY.** These are generator refusals, not
executor availability decisions, and therefore this evidence does not claim
the capability generalizes or is ready to merge. The complete evidence is in
`.planning/phases/persistiq-artifact-materialize-pilot-r1/generalization-validation-2026-08-08/`.
The eligible 392 remain untouched and are deferred. Certification remains
withheld; no provider operation was exercised.

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
