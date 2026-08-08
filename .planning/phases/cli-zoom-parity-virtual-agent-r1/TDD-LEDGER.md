# TDD Ledger — Zoom Virtual Agent documented-operation parity, R1

## Planned RED contract

Before any Virtual Agent production declaration changes, the RED checkpoint contains only the
command-surface test and phase evidence. Against SCIM2-complete HEAD it must fail because:

- Zoom is at `38` executable / `1,804` locally implementable rows, with `21` direct reads and
  `12` direct writes; the Virtual Agent target is `51` / `1,791` / `30` / `16`.
- All thirteen provider paths are absent from the real commandrunner preflight, so a compiled
  `pm zoom virtual-agent …` route remains an `unknown command` before its declaration exists.
- No declaration can turn response-only `page_size`, `next_page_token`, `from`, `to`, or other
  response schema values into an invented request flag.

The RED output will be pasted verbatim below before any production JSON, metadata, fixture, or
generated-file edit.

## RED — pending

## GREEN connector — pending

## Verification/review — pending
