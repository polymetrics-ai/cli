# TDD Ledger — #3858 page-safe polling source executor

## Slice 1 — preflight-resolved bounded page execution

- Red: pending — create the source/sink/store test that must observe one
  catalog-bound tuple request, one durable page acknowledgement, and a #3810
  committed tuple. The current preflight source interface contains no read
  method, so there is no page-safe source path to exercise.
- Green: pending — add the smallest closed executor path; record the focused
  `go test -timeout 20m` output and exact observable assertions here.
- Refactor: pending — defensive clone, context, and typed-error review.

## Slice 2 — interruption and durable-after-ack

- Red: pending — page-split fixture stops after page one and an
  acknowledgement-success/state-save-failure run; both must show the old
  committed tuple is resumed and no checkpoint is advanced by an attempt.
- Green: pending — combined delivery is exactly `a`, `b`, `c`, with a
  checkpoint only after the full page acknowledgement and successful store.
- Refactor: pending — audit ordering and bounded-memory behavior.

## Slice 3 — fail-closed resume and cursor policy

- Red: pending — inject null/precision, non-advancing tuple, source/schema
  mismatch, and unsafe-overlap cases. Each must prove zero reads, zero durable
  destination acknowledgements, and zero checkpoint-store writes as relevant.
- Green: pending — typed recovery/refusal outcomes and immutable corpus
  observations pass through the real executor.
- Refactor: pending — verify no scalar cursor, float64, display formatting,
  or raw protocol input is reachable.

## Status

Planning checkpoint created before production edits. No red test or production
implementation has been written yet.
