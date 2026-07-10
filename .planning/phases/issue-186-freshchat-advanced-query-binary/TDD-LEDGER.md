# TDD Ledger — Issue #186

## GSD note

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`), so this slice uses the manual GSD/TDD fallback.

## Red target

- Freshchat embedded bundle must expose typed `file_upload` operations for `freshchat.files.upload` and `freshchat.images.upload`.
- `pm freshchat file upload` / `pm freshchat image upload` must be operation-backed and fail closed before credential resolution.
- Connectorgen should validate that Freshchat binary upload commands use typed operation metadata instead of a raw upload escape hatch.

Expected initial failure: Freshchat has no `operations.json`, and upload commands are merely excluded metadata entries with no typed operation IDs.

## Green target

Add operation metadata and command-surface wiring while preserving blocked-by-default upload behavior.

## Verification ledger

Pending.
