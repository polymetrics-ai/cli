# CONTEXT — durable fencing and parking production wiring

## Scope

Close the identical re-audit residuals on #3865 and #3867: replace the
process-only persistence seams with durable stores and make both coordinators
reachable from `app.Open`, normal admission, credential test/repair, ETL park,
and restart/resume paths.

## Locked decisions

1. Persist coordination truth with the repository's existing atomic,
   file-locked JSON state mechanism under `.polymetrics/state`. Do not add a
   database, cache, migration framework, or new dependency to the default path.
2. Keep only secret-free opaque `AuthCohortKey` and `RateLimitScopeKey` values,
   typed outcomes, committed checkpoints, reset evidence, and schema version in
   coordination files. Never persist credential material, raw provider values,
   response bodies, URLs, or headers.
3. `app.Open` constructs both durable stores and coordinators. Credential
   resolution supplies ordinary auth admission to the connector runtime;
   `pm credentials test` is the dedicated probe that can fence on a typed
   verified-invalid result or advance a healthy epoch on verified success.
4. Normal connector operations join a cancellable auth admission before I/O.
   Connector-specific code may mark a failure verified only from a typed,
   authoritative condition (for PostgreSQL, SQLSTATE 28P01); an ordinary 401,
   timeout, transport error, or provider failure cannot fence.
5. Rate parking admission is composed with the declaration-derived opaque rate
   scope before each send. ETL parks only a typed terminal rate-limit error with
   authoritative reset evidence and an already committed checkpoint.
6. Restart reloads parked work. Work already due is attempted synchronously
   during `app.Open`; future work is timer-backed while the process remains
   alive. Resume re-enters `App.RunETL` from durable stream state and suppresses
   nested parking so a failed retry leaves the original record authoritative.
7. File-store mutations are cross-process atomic. Parking claims use a bounded
   durable lease so two restarted processes cannot both resume, while a process
   killed after claiming becomes retryable after lease expiry.
8. Tests must spawn and terminate real child processes. Fencing is exercised
   through the shipped CLI against a real containerized PostgreSQL server.
   Parking uses the shipped CLI/application composition and real filesystem/HTTP
   I/O with controlled fault responses, then observes rows/checkpoints and store
   state after a second process starts.

## Explicit exclusions

- Do not fix #4125, #4136, or #4158.
- Do not change a connector capability flag or add a public command/flag.
- Do not use GitHub credentials or a generic HTTP/SQL/shell write surface.
- Do not weaken reverse-ETL approval, target durability, or checkpoint ordering.

## GSD fallback

`scripts/gsd doctor`, command-source resolution, and
`go run ./cmd/agentcontractgen check` passed. Generated `discuss-phase` and
`plan-phase --tdd` prompts are executed inline because the canonical
single-worker contract forbids role spawning and this issue pair is not a
numbered roadmap phase. The fallback does not waive TDD, verification, live
evidence, or review.
