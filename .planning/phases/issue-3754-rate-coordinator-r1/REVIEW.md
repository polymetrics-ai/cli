# Scoped code review — issue #3754

Scope: all changed coordination, connector engine/SDK, CLI, generated documentation, migration
conventions, and website reference files. Comparison base: `integration/4015-mvp-flat-r1` at
`fbd06e7d7c5c0632182e98cbb3a223ba25b19883`. Files outside this diff are out of scope.

## Result

No blocking or warning findings.

- Shared registry selection is solely the closed `coordination: require_shared` declaration; a
  configured coordinator never upgrades an absent/local policy.
- Missing or unreachable coordination returns `SharedRateLimitUnavailableError`; resolver setup
  fails before creating/sending a shared requester and has no local fallback.
- Shared admission has one opaque key and one Redis Lua transaction using server `TIME` plus TTL.
  Waiting observes the caller context; two real child processes demonstrate atomic grant/block.
- Key construction consumes only the existing #3863 `RateLimitScopeKey` projection. Error, status,
  CLI, and test helper paths avoid raw credentials, scope subjects, binding values, endpoint values,
  or environment inheritance.
- No production connector bundle, connector-specific execution branch, persistence/receipt, parking,
  resumption, or generic execution surface was introduced.

## Review validation

Focused package/race tests, live Dragonfly integration proof, vet/build/diff check, CLI parity, and
all individual repository gates listed in `VERIFICATION.md` passed.

## Correction 3/5 review scope — #4035

Review only the reservation lease lifecycle and the fixed UDS protocol. Required dispositions: no raw scope, endpoint, request, or credential is emitted; no generic RPC is added; expiry frees concurrency but not a valid completion observation; cancellation aborts read/write I/O and wins any post-cancellation response race. Findings outside `internal/coordination/**`, `internal/connectors/connsdk/rate_budget_coordinator.go`, and this correction evidence are `needs-decision` follow-ups, not fixes in this child.

Review completed: no findings. The private protocol has exactly ready, decide, and finish message kinds; the client derives an I/O deadline from the caller context, advances it using `context.AfterFunc`, and checks `ctx.Err()` after reading a response. The only state retained past lease expiry is the opaque lease-to-budget mapping necessary to apply a valid late completion; it no longer occupies a concurrency slot. `go vet`, repository `make lint`, formatting, `git diff --check`, focused engine checks, and the required race checks passed. The known `window_seconds` duration-overflow defect in `internal/coordination/shared_rate_limits.go` was not changed (tracked as #4125).

## Correction 5/5 review scope — #4049

Scope: `internal/connectors/engine/rate_limit_runtime.go`, GitHub WriteHook
request routing, their deterministic local tests, and this delivery evidence.
Comparison base is `integration/4015-mvp-flat-r1`. No GitHub rate-limit
declaration, credential, CLI/docs surface, provider call, generic HTTP writer,
or unrelated coordinator behavior is in scope.

Review completed: no findings. The engine retains the base requester only for
whole-connector policies; when a config-matching selector needs a method/path,
the default requester has a guard that returns before either transport or
generic admission. `Runtime.RequesterFor` replaces that guard and retains the
normal route resolver at each physical send, including redirects and escaped
paths. All nine WriteHook sends use one small declaration-aware helper with
bundle-declared constants; compound follow-ups acquire their own requester.
Tests assert zero local sends for both direct and `require_shared` refusals,
and a table test asserts positive expected send counts for all fourteen
declared hook paths. `rg` found
no remaining `rt.Requester.Do` call in the GitHub hook; `rate_limits.json` has
no diff. The known `window_seconds` overflow in
`internal/coordination/shared_rate_limits.go` remains untouched (#4125).
