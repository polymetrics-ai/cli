# Overview

When Clause Equality Unknown Spec Key is a deliberately invalid connectorgen validate corpus case
(S3 engine mini-wave item 2): its `streams.json` `base.auth`'s first candidate's `when` clause uses
the `==` when-grammar operator (`{{ config.auth_typo == 'token' }}`) referencing an undeclared spec
key (`auth_typo`, a typo of `auth_type`), which `engine.ResolveCheckWhen` (wired into
`engine.ResolveCheckAuthSpec`'s `when` field check) must reject. The second candidate's `when` uses
the `in` operator against the correctly-declared `auth_type` key, proving the fixture isolates the
`==`-typo failure specifically (not a blanket rejection of the whole when-grammar).

## Auth setup

Bearer/none dual candidates, deliberately misconfigured for this test case.

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

None; read-only test fixture.

## Known limits

None; this is test fixture data.
