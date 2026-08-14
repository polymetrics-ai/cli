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
