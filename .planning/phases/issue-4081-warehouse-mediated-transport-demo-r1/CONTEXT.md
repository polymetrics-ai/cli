# #4081 — warehouse-mediated Transport demo context

**Issue:** [#4081](https://github.com/polymetrics-ai/cli/issues/4081)
**Parent:** [#4015](https://github.com/polymetrics-ai/cli/issues/4015)
**Phase depth:** normal; one issue-sized walking slice
**Mode:** `discuss-phase --auto` inline/manual fallback

## Locked decisions

1. This is one GitHub → warehouse → GitHub walking slice, not the four-family
   certification matrix. PostgreSQL legs, schedules, flows, CDC, rate/auth,
   query, broad mode certification, generator promotion, and Shepherd are
   explicitly deferred.
2. The only route is closed GitHub source → connection-owned durable warehouse
   stage → independent `Reopen(handle)` → typed GitHub destination. There is no
   generic HTTP/SQL/shell write, raw provider cursor, raw-page apply, or
   connector-to-connector call.
3. `Stage` returns only after the existing owner-scoped JSONL WAL, DuckDB
   materialization, Parquet publish, and directory durability boundary succeeds.
   The production handoff is an immutable handle/receipt, never a source slice.
4. `Reopen` must validate owner, generation, manifest, and content identity
   before bounded records are reconstructed from durable storage. The demo drops
   all original source-page and staged record references before reopen.
5. Destination work remains on the existing typed plan → preview → approval →
   execute path. A durable receipt plus independent read-back is required before
   checkpoint CAS; destination/read-back/checkpoint failures retain safe replay
   and cannot advance state.
6. The local proof uses a faithful GitHub test server and uniquely run-owned
   fixture resources. A real provider run is optional and only permitted through
   PM's approved GitHub App credential path and a disposable `Polymetrics-Cert`
   resource. Credential failure is a fixed safe blocker, never a fallback to a
   PAT, personal repository, or secret inspection.
7. #4079's corrected record-isolation parent is a hard dependency. The demo
   must exercise the closed mutable-value boundary without weakening it.
8. The artifact is a bounded exact-binary faithful-server test plus the accepted
   closed `pm etl transport github-issue-label` plan/preview/cleanup lifecycle
   and its narrowly extended approval-carrying `pm etl run`. The carrier accepts
   only connection, closed plan selector, stdin marker, and typed destructive
   confirmation; every provider detail stays connection/App-owned. CLI
   help/manual/website parity is part of the same TDD slice.

## Base admission

The authoritative remote `docs/4015-connector-release-certification` head is
`e7d2b2963fc1dd164f63b31fccb8a3bab8084bec`. PR #4019 was squash-merged, so the
accepted #4077 source commit `aaf288d069adc1b67a09500afcca4be4a6d1bab3` is not
required to be an ancestor. Instead, the required Transport implementation and
regression-test blobs at the combined head were compared directly and are
identical to `aaf288d...`; its #4077 phase evidence is also present in the tree.

The child branch is therefore created from exactly `e7d2b296...`. The planning
checkpoint remains the first commit on that branch; only after it is committed
may RED tests and production work begin.

## GSD inline/manual fallback

The named issue phase is not a numeric roadmap entry and this Codex environment
cannot provide the compatible isolated Pi roles expected by upstream GSD. The
repository's canonical single-worker contract also forbids role spawning.
`scripts/gsd prompt` resolved the named official commands; this phase performs
their required reasoning and records durable artifacts inline.

## Required skills used

- `github-issue-first-delivery`
- `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`
- `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`,
  `golang-concurrency`, `golang-database`, `golang-testing`, `golang-documentation`,
  `golang-lint`
- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`,
  `gsd-code-review`, and `no-mistakes`

## Deferred items

- PostgreSQL source/destination adapters and all three other compositions.
- Schedules, flows, bounded database query, CDC, restart orchestration beyond
  the named fault injections, cross-process auth/rate coordination, #3865,
  #3867, #3866, #3994, #3992, #3993/#3996, and the final #4016 promotion.
- Generic certificate matrix/generator truth changes and GitHub
  `full_overwrite` resource semantics.
