# Overview

Test-only bundle for the shared polling-watermark executor.

## Auth setup

None; no live provider is contacted.

## Streams notes

`widgets` is ordered by `updated_at`, then `id`.

## Write actions & risks

None.

## Known limits

Fixture only. Its declared soft-delete field makes the executor emit a tombstone;
hard-delete visibility is intentionally not represented here.
