# Overview

Acme GitHub Range Cursor Offset is a synthetic connector used as a conformance v2 self-test
bundle for the `github_date_range` `param_format` (N1, wave0 REVIEW.md re-review) fed a STRING
cursor value carrying a non-UTC RFC3339 offset (`+05:30`), rather than a bare digit string.
Locks in that `formatCursorForAssertion`'s `github_date_range` branch normalizes a non-UTC
RFC3339 cursor to UTC second precision exactly like the engine's `formatParam` does, instead of
returning `">=" + value` VERBATIM (which would assert against the un-normalized offset form and
never match what the real engine actually sends on the wire).

## Auth setup

No auth required; public synthetic API.

## Streams notes

`events` is incremental on `updated_at` (an RFC3339 string) and has no pagination.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
