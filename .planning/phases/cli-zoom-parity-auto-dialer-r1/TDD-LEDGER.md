# TDD Ledger — Zoom Auto Dialer documented-operation parity, R1

## Planned RED contract

Before any Auto Dialer production declaration changes, the RED checkpoint contains only the
command-surface test and phase evidence. Against Virtual-Agent-complete HEAD it must fail because:

- Zoom is at `51` executable / `1,791` locally implementable rows, with `30` direct reads and
  `16` direct writes; the Auto Dialer target is `67` / `1,775` / `38` / `24`.
- All sixteen provider paths are absent from the real commandrunner preflight, so a compiled
  `pm zoom auto-dialer …` route remains an `unknown command` before its declaration exists.
- The existing named-root-object support is fixed-operation and schema-bound; RED does not add a
  generic body, HTTP, or paging capability.

The RED output will be pasted verbatim below before any production JSON, metadata, fixture, or
generated-file edit.

## RED — pending

## GREEN connector — pending

## Verification/review — pending
