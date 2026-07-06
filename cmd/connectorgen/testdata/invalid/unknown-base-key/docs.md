# Overview

Unknown Base Key is a deliberately invalid connectorgen validate corpus case (ENGINE HARDENING,
hardening-ledger.md; updated by checkquery-ledger.md when `base.check.query` was legitimized): its
`streams.json` `base` declares a bare top-level `query` field (`{"limit": "1"}`) — a per-bundle
"shared query params applied to every stream" mechanism that has never existed on `HTTPBase`, unlike
`base.check.query` (now a real `RequestSpec.Query` field) or `stream.query` (an existing,
per-stream field). This reproduces the still-unrepaired 7-bundle `base.query` class named in
hardening-ledger.md's "Newly exposed" section (reply-io, retailexpress-by-maropost, retently,
revenuecat, revolut-merchant, ringcentral, you-need-a-budget-ynab), which remains out of scope for
checkquery-ledger.md (that dispatch covers `base.check.query` only). Both the streams.schema.json
meta-schema (`base` declares an explicit property allowlist with `additionalProperties: false`, and
`query` is not on it) and the loader's independent `strictDecode` pass must reject this.

## Auth setup

Bearer token, not exercised by this test case (the defect is structural, at load time, before any
auth/read path runs).

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

None; this bundle has no writes.json.

## Known limits

None; this is test fixture data.
