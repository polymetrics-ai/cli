# TDD ledger — Gmail parity wave03

## Red / failing evidence before production edits

- Official discovery re-audit artifact has 79 operations: `.planning/issues/gmail-parity-wave03/official-gmail-discovery-operations.json`.
- Pre-change Gmail surface classified 34 operations as `excluded` and had no `operation` rows:
  - covered.stream=10
  - covered.write=35
  - excluded.binary_payload=2
  - excluded.destructive_admin=1
  - excluded.duplicate_of=15
  - excluded.non_data_endpoint=2
  - excluded.requires_elevated_scope=14
- This failed #3046/#3047 because only 11 CSE operations were true excluded/N/A; the other 23 needed executable coverage or blocked operation rows with exact evidence.
- This failed #3050 because Gmail documents two batch message mutations not represented as typed writes.
- This failed #3049/#3051 because Gmail had no definition-owned provider CLI/direct-read surface for bounded detail/binary-like reads.
- The documented focused validation gate `go run ./cmd/connectorgen validate internal/connectors/defs/gmail` failed by validating child directories (`fixtures/`, `schemas/`) as connectors instead of the single bundle.

## Green evidence added

- `api_surface.json` now uses `operation_ledger_version: 1` and enumerates all 79 Gmail Discovery revision `20260727` operations exactly once.
- Post-change counts: total=79; executable=61 (10 streams + 11 direct reads + 40 writes); blocked/planned=18 (5 email-path direct reads + 2 CDC controls + 11 CSE admin/N/A rows); certified=0.
- New write actions have closed top-level record schemas and sanitized fixtures:
  - `batch_modify_messages`
  - `batch_delete_messages`
  - `insert_smime_info`
  - `set_default_smime_info`
  - `delete_smime_info`
- Implemented direct-read commands are bounded, fixed-target, operation-backed, and redacted:
  - `messages get`, `threads get`, `drafts get`, `labels get`, `filters get`
  - `settings auto-forwarding`, `settings vacation`, `settings language`, `settings imap`, `settings pop`
  - `attachments get`
- Email-address path direct reads (`send-as get`, `delegates get`, `forwarding get`, `smime list`, `smime get`) are planned/blocked, not advertised executable, because the shared operation direct-read path-variable validator currently accepts identifier-safe values only.
- Generated Gmail docs/SKILL, connector README/catalogs, website connector data, and CLI golden transcripts reflect the new surface.
- `cmd/connectorgen validate` now accepts either a defs root or a single connector bundle directory; regression test added.

## Refactor / cleanup notes

- Production Gmail behavior changes are definition-local under `internal/connectors/defs/gmail/**`.
- Non-Gmail generated connector MANUAL/SKILL drift from full doc generation was intentionally restored to keep the branch scoped.
- No shared engine/runtime behavior changes were made.
