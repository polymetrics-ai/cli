# Reddit full-parity completion context

## Source decisions

- The verified 2026-08-04 Reddit API ledger is the work list; it is not to be re-extracted.
- The 2026-08-10 captain resume order supersedes the earlier narrow-scope brief for this branch.
- The rebased checkpoint reports 230 documented `api_surface` rows, 221 covered rows, and 220 implemented commands. The remaining nine rows must each reach either executable coverage or a named, machine-checkable exclusion.

## Locked implementation decisions

1. Implement the three S3-lease upload operations after first checking the existing `file_upload` executor for the required acquire-lease then bounded third-party multipart PUT flow.
2. Implement `pm reddit vote` only as a human-invoked, direct interactive command with a per-invocation confirmation. It must not be batchable, schedulable, or reachable through reverse-ETL planning.
3. Retain five exclusions with exact named reasons: Reddit Premium (`store_visits`), approved OAuth app (`block_user`), deprecated endpoint (`recommend/sr`), legacy captcha flow (`needs_captcha`), and superseded documented per-article comments route.
4. Keep the connector free of `unsafe_or_disallowed` dispositions. Moderator and live-thread actions remain executable where the token has provider-granted authorization.
5. Use no credentialed or live Reddit calls. Reverse-ETL checks remain plan -> preview -> approval -> execute; no operation is executed during validation.
6. Regenerate connector catalogs instead of hand-merging generated output after the rebase.

## Scope and safety

- Primary connector ownership is `internal/connectors/defs/reddit/`, its fixtures and schemas, the Reddit hooks already on this branch, planning evidence, and connector-generated catalogs.
- The captain explicitly permits a missing bounded-transfer or non-batchable foundation to land inline in this issue if the existing engine cannot express it. Any shared change will be isolated in its own commit and disclosed with its reusable impact.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-documentation`.

