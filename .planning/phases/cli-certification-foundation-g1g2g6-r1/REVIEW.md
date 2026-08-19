# Code review: connector certification foundation G1/G2/G6

## Method

Manual standard-depth review under the documented GSD inline fallback. Reviewed
the parity classifier and sweep validator, staged evidence publisher and every
batch caller, draft-import boundary, scoped matrix check, live-runner ordering,
generated GitHub artifact, tests, and architecture documentation. Re-ran
`go test -timeout 20m ./cmd/connectorgen`, `go vet ./...`, `make lint`, and
the generated-artifact checks after review.

After rebasing onto `origin/integration/4015-mvp-flat-r1`, repeated the
review for the declaration-row reconciliation: a native bundle without
`cli_surface.json` now remains sweepable, capability/changefeed/transport rows
have stable non-CLI identities, synthetic rows cannot borrow generic
`write:false`, and the totals distinguish CLI commands from all declarations.
Re-ran focused projection, artifact, full generator, full-suite, generated,
lint, docs, smoke, boundary, release, and live-GitHub proof checks on that base.

## Findings

No critical, warning, or actionable informational findings.

The publisher stages complete bytes in the destination directory, syncs before
the hard-link no-replace publication, removes staging names on every exit path,
and validates every planned destination before staging. A concurrent writer can
still win a final-name race, correctly receiving a no-replace error; no reader
can observe a partial final JSON record. Multi-record callers prepare all
records before publication, so validation/destination-precheck failures emit no
prefix.

The scoped check intentionally does not read status.json or another connector's
shard. The existing global `--check` retains its complete status/shard
validation and remains final-fan-in-only.

## Review disposition

Clean. Automated GitHub review is pending PR creation and is recorded in the
PR body as `claude_auto` / stacked-parent fallback as applicable.
