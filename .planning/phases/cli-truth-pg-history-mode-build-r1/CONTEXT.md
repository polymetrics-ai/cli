# Context — PostgreSQL history-mode truthfulness repair

## Decision

Make the existing PostgreSQL `incremental_dedupe_history` claim true. Do not
remove, downgrade, or qualify the published capability.

The defect is descriptor drift: the inner PostgreSQL polling declaration and
database writer already admit the mode and `dedupe_history` strategy, while
`internal/connectors/defs/postgres/sync_transport.json` does not. The outer
descriptor is the production registry's authoritative admission point.

## Defined history semantics

The existing typed contracts fully determine the behavior; no product decision
is outstanding.

- A history row is one mapped source record plus PostgreSQL's `_valid_from`,
  `_valid_to`, and `_is_current` fields. `internal/connectors/native/postgres/
  managed_target_write.go` creates the layout only for this mode.
- The connection's validated `--primary-key` fields identify the lineage. A
  newly observed value for that key closes the prior current row at the new
  row's `_valid_from`, marks it non-current, and inserts the new current row in
  the same PostgreSQL transaction.
- The source's validated `--cursor` plus primary key form the ordered polling
  tuple. The durable workset carries its source-owned final page checkpoint,
  which the PostgreSQL destination attaches to each record in that bounded
  page. Keyset traversal makes a later page's tuple later than the prior
  committed page; the keyed order fence accepts only later tuples. Equal or
  late replays are no-ops and cannot reopen a closed history window. An
  identical observed version is also a no-op.
- A delete tombstone closes the current validity window and retains all prior
  history; omission from a later source page is not a delete.
- The destination write session publishes its bounded batches and receipt
  before the source checkpoint/acknowledgement advances. This preserves PR
  #4184's run-scoped publish-then-checkpoint atomicity; this task introduces no
  per-page checkpoint.

## Scope

Exactly one target connector is changed: `postgres`. Production changes are
limited to its declarative outer transport descriptor and native PostgreSQL
source/destination adapters: the source runners carry their already-loaded
source definition into the existing inner history-route seal; the destination
seals the outer route, preserves the declared primary key, and attaches the
durable source position to history records. App/CLI tests assert the descriptor
reaches the existing shared registry and production binary; shared runtime
behavior is not modified.

## Compatibility

No new command, flag, help text, manual text, or website prose is required:
the advertised PostgreSQL docs and generated website payload already describe
the mode accurately. Generator and repeat-generation checks must prove they
remain byte-stable.
