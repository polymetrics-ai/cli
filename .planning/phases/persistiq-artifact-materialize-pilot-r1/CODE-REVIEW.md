# PersistIQ artifact materialization pilot - code review

**Status:** clean for the captain-policy implementation and PersistIQ pilot
evidence; certification remains withheld

**Review mode:** manual inline fallback. The pilot phase is not present in
`ROADMAP.md`, and this Codex runtime cannot provide the interactive Pi worker
required by the installed GSD adapter. The fallback and its Red/Green evidence
are recorded in `CONTEXT.md`, `DISCUSSION-LOG.md`, `PLAN.md`, and
`TDD-LEDGER.md`.

## Scope reviewed

- The policy-change commits on `fm/cli-mass-artifact-materialize-r1`.
- Complete operation inventory materialization and named-dependency metadata.
- The exact source-surface discrepancy marker for PersistIQ's three
  undocumented legacy streams.
- The deferred single final gate design: fetch, materialize, then gate the
  staged result once.
- Fresh PersistIQ artifact bytes/hash/parse evidence, generated bundle,
  static reports, timings, and real-binary sweep.

## Findings

No correctness, security, or scope findings were identified.

The old coverage refusal was removed only for inventory completeness. Missing
executor/foundation support is represented as `not_implemented` with a
machine-checkable named dependency. The runtime-preflight invariant remains
unchanged: the generated 21 implemented commands passed the production
`commandrunner.Preflight`, while all three unsupported commands were visible
and blocked before credentials/network.

The three existing source-only rows are retained, not silently deleted, and
carry `present-in-surface-absent-from-artifact`. The generated artifact has 21
documented operations plus those three discrepancy rows, for 24 surface rows.

## Verification

- `go test -timeout 20m ./cmd/connectorgen` passed.
- `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner` passed.
- `go vet ./cmd/connectorgen` passed.
- `go build ./cmd/pm` passed.
- `go run ./cmd/agentcontractgen check` passed.
- Generated `connectorgen validate` returned 0 findings.
- Generated `surface-sync --check` reported no drift.
- Generated `connectorgen batch gate` included PersistIQ with 21 runtime-
  preflight checks and dropped 0 candidates.
- `TestEveryImplementedCommandPassesRuntimePreflight` passed.
- The real built `pm` binary reached 24/24 generated help paths, with 0
  unknown-command failures; `pm persistiq --help` also exited 0.
- `git diff --check` passed and the source PersistIQ bundle was restored after
  the embedded-binary test; no production connector definition was left
  modified.

This review does not certify PersistIQ. The connector is implemented according
to static evidence, not certified, and never exercised against its provider.
