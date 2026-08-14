# TDD Ledger — #3858 page-safe polling source executor

## Slice 1 — preflight-resolved bounded page execution

- Red: `go test -count=1 -timeout 20m ./internal/connectors/engine -run
  '^TestPollingSourceExecutor'` failed to build because
  `PollingSourceExecutor`, `PollingSourceRuntimeState`,
  `PollingSourcePageRequest`, `PollingSourcePage`, and `PollingSourceItem`
  did not exist. The test already asserted catalog-bound tuple requests,
  durable page acknowledgement, and a #3810 committed tuple, so a no-op
  implementation could not pass.
- Green: the same focused command passes after adding the closed executor.
  The fake runner must consume the request budget before every fetch; the fake
  destination records delivered row IDs and calls #3810's real
  `CommitAfterDownstreamAcknowledgement`; the store records actual committed
  envelopes.
- Refactor: runtime state and request inputs defensively clone opaque tokens;
  every tuple comparison is byte-for-byte and no source request includes raw
  protocol input.

## Slice 2 — interruption and durable-after-ack

- Red: page-split and acknowledgement-success/state-save-failure fixtures
  were added before the executor existed. They observe exact tuple requests,
  delivered stable identities, and persisted envelopes rather than merely an
  exit status.
- Green: combined logical delivery across an interruption is exactly `a`,
  `b`, `c`; the second first-run request is `(watermark,b)`, proving the next
  page is requested only after page one committed. A failed state save leaves
  zero new envelopes, and the next invocation requests the old committed tuple
  and replays `b`. An empty page observes one fetch, zero destination sends,
  and zero checkpoint writes.
- Refactor: source sequencing updates its in-memory `after` tuple only after
  the page emitter returns successfully. The shared transport path is the
  durable sink boundary, so stage/apply/read-back/commit failures all preserve
  the prior tuple.

## Slice 3 — fail-closed resume and cursor policy

- Red: tests inject schema mismatch, source-generation mismatch, null cursor,
  and a repeated complete tuple. The test records fetches, page delivery, and
  checkpoint-store calls; a refusal must show the declared zero side effects.
- Green: stale resume state is refused before `FetchPollingSourcePage`;
  malformed/repeated page tuples are refused before the destination or store.
  The real executor also forwards a declared soft-delete as a validated
  tombstone and commits its final page tuple only after that page is durable.
- Refactor: timestamp tokens require canonical `RFC3339Nano`; decimal tokens
  are validated by `math/big.Int` without converting the opaque checkpoint
  bytes. Float/scalar fallbacks, offsets, raw query inputs, and cursor-only
  resume paths are absent.

## Status

The first implementation pass passed the focused real-executor tests. Full
package/regression, vet, repository gates, review, rebase, and PR checks remain
before delivery.
