# Overview

When Clause Equality Valid is a control connectorgen validate corpus case (S3 engine mini-wave item
2, wave1-pilot SUMMARY.md carried queue / REVIEW-A.md re-review R1/R3): its `streams.json`
`base.auth`'s two candidates gate on `==` and `in` when-grammar comparisons against the spec-declared
`auth_type` key (`{{ config.auth_type == 'token' }}`, `{{ config.auth_type in ['public', 'none',
'anonymous'] }}`) — exactly the shape `engine.ResolveCheck`'s prior bare-namespace.key-only parsing
could never statically validate (it hard-failed with an "unknown spec key" finding even though
`auth_type` IS declared, because it treated the whole `auth_type == 'token'` expression as one dotted
reference). `engine.ResolveCheckWhen` (wired into `ResolveCheckAuthSpec`'s `when` field check) now
parses the comparison operators and validates only the left-hand-side reference, so this bundle passes
`connectorgen validate` with zero findings.

## Auth setup

Bearer (token mode) / none (public mode) dual candidates, gated by an explicit `auth_type` enum —
this is the canonical rendering of a legacy `auth_type`-string-enum auth surface that github's
`public_access` boolean-opt-in workaround (ledger G14) can now retire in favor of, once this fix
lands.

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

None; read-only test fixture.

## Known limits

None; this is test fixture data.
