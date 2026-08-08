# TDD Ledger — Zoom Workforce Management documented-operation parity, R1

## Planned RED contract

Before any Workforce Management connector declaration or CSV-foundation production change, the
test-only RED checkpoint must prove both gaps:

- Tasks-complete HEAD has `84` executable / `1,758` Zoom-local rows, `44` direct reads, and `35`
  direct writes; this category's target is `102` / `1,740` / `55` / `42`.
- All 18 real `workforce-management …` paths are unknown to the command runner before their
  declarations exist.
- A closed multipart `content_validation: "csv"` declaration is rejected because the existing
  policy supports only JSON validation; a `.csv` filename alone is not adequate source validation.

RED output will be recorded verbatim below before production changes.

## RED — pending

## GREEN CSV foundation — pending

## GREEN connector — pending

## Verification/review — pending
