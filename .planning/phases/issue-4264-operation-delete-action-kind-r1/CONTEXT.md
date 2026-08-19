# Issue 4264 — operation-backed mutation action kinds

## Fixed decisions

- This is a shared `cmd/connectorgen` foundation fix. It must not add
  connector-specific classifier branches or modify Zoom definitions.
- `write_action_kind` is generator-owned: it is derived from a linked
  `writes.json` declaration when present, otherwise from the linked operation
  surface. A mutation with no determinable action kind is an error.
- Delete selection is proven from generated GitHub and Zoom sweep artifacts,
  with no credentials or live provider mutation.
- Existing `writes.json` classifications remain authoritative and continue to
  be covered independently.

## Delivery constraints

- Issue: Refs #4264. Base and PR target: `main`; final destination: `main`.
- Branch: `fm/cli-delete-action-kind-fix-r1`.
- Regenerate only generator-owned sweep artifacts affected by the new derived
  field. Keep the focused diff below the supervisor's 15-file threshold.
- The GSD adapter's Pi role runtime is unavailable in this environment, so the
  approved inline/manual lifecycle fallback is used and recorded in this phase.
