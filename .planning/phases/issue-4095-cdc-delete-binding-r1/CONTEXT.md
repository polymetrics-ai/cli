# Context — Issue 4095: PostgreSQL CDC delete binding

## Task Delivery Header

- Issue: Closes #4095 — feat(postgres): bind change-capture deletes to keyed apply and history close
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green.
- Working branch: fm/cli-4095-cdc-delete-binding-r1
- Task: Convert an explicit PostgreSQL CDC delete event into a source-keyed tombstone, map only those declared keys through `MappingContractV1`, and deliver it through the existing keyed PostgreSQL history-close path. An omitted record remains an ordinary absence, never a delete.
- Verification: Targeted database/PostgreSQL tests, the tagged Colima Docker dbtest run, focused static/repository gates, and API read-back of the opened PR base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An explicit CDC delete closes, rather than physically removes, the current history row | live | The tagged PostgreSQL test reads the actual managed target and asserts the last version remains stored, has `_is_current=false`, and has a non-null `_valid_to` after delivery of a CDC-derived tombstone. |
| Physical absence remains non-destructive while an explicit CDC tombstone applies | live | The tagged PostgreSQL workset test reads the target after an absent source row and observes it still present; after the explicit CDC-derived tombstone, the same key is absent from the non-history target. |
| CDC delete keys traverse the sealed mapping contract | fake | A deterministic unit test checks a source-keyed CDC tombstone maps only the declared source keys to their target names; malformed/missing keys are rejected before a write input can be constructed. The real target test separately proves the resulting write effect. |

## Locked decisions

- The four named foundations exist on the rebased integration base: pgoutput-v2 CDC, managed PostgreSQL writing, `MappingContractV1`/`DeliveryReceiptV1`, and `incremental_dedupe_history`.
- Keep `change_capture` PostgreSQL-source-only. The target remains the already-admitted keyed apply mode (`incremental_dedupe_history` here); this task adds no destination mode or public command.
- Treat only a CDC event whose operation is `delete` as a tombstone. A row omitted from a batch has no delete meaning.
- Map source key names with the shared mapping contract once, before the direct keyed write path. Reuse the same mapping operation for immutable worksets so the two delete routes cannot drift.
- Use the Postgres CDC LSN plus canonical key payload for a deterministic opaque tombstone identity and source-owned position. Do not expose a raw LSN-writing surface.
- Inline/manual GSD is necessary: this direct issue phase is non-numeric and compatible isolated GSD workers are unavailable. The command prompts were resolved with `scripts/gsd prompt ... --auto`; no lifecycle or TDD gate is waived.
