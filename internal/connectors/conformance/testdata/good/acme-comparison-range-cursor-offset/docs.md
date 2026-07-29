# Overview

Acme Comparison Range Cursor Offset is a synthetic connector used as a conformance v2 self-test
bundle for generic `rfc3339_utc` plus `operator_prefix` incremental formatting fed a STRING cursor
with a non-UTC offset. It proves conformance normalizes the expected request parameter the same way
as the engine: parse the offset timestamp, emit UTC RFC3339, then prepend the comparison operator.

## Auth setup

No auth required; public synthetic API.

## Streams notes

`events` is incremental on `updated_at` and has no pagination.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
