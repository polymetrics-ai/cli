# Overview

Acme Numeric Cursor is a synthetic connector used as a conformance v2 self-test bundle whose
incremental cursor field is a JSON NUMBER on the wire (mirroring Stripe's `created` Unix-seconds
field), rather than a string, to lock in numeric-cursor support in `cursor_advances`.

## Auth setup

No auth required; public synthetic API.

## Streams notes

`events` is incremental on `created` (a Unix-seconds integer) and has no pagination.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
