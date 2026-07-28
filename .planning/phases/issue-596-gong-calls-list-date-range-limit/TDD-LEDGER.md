# TDD Ledger — issue 596 Gong calls list correction

## Red

- [ ] Add fail-first tests for `calls list --from` / `--to` request shape.
- [ ] Add fail-first tests for invalid timestamp and invalid range rejection before HTTP.
- [ ] Add fail-first tests for `--config start_date` compatibility and explicit `--from` precedence.
- [ ] Add fail-first tests for `--limit` counts across cursor pages: 1, below boundary, boundary, above boundary.
- [ ] Add fail-first/help assertions for `pm gong calls list --help` command-specific flags and output-cap wording.

## Green

- [ ] Implement command flags, validation, query precedence, and docs.
- [ ] Targeted tests pass.

## Refactor

- [ ] Keep scope narrow; no participant stream, live checks, or unrelated Gong endpoint edits.
- [ ] Re-run formatting and validation after docs generation.

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `no-mistakes`.
