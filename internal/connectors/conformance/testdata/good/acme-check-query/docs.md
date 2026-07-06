# Overview

Acme Check Query is a synthetic connector used as a conformance v2 self-test bundle for
`base.check.query` (`RequestSpec.Query`, checkquery-ledger.md): its `fixtures/check.json` records a
`request.query` that matches exactly what `engine.Check()` sends, so `check_fixture` must pass.

## Auth setup

No auth required; public synthetic API.

## Streams notes

`events` is a single stream, no pagination, no incremental.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
