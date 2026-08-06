# Manual code review — Issue 3810

**Scope reviewed:** `internal/synccontract/**`, `internal/app/stream_state.go`, and the targeted
state/ETL compatibility changes in `internal/app`.

## Findings

No unresolved Critical, Warning, or Info findings.

## Checks performed

- Confirmed the only shared vocabulary is `synccontract.Mode`; legacy app spellings are explicitly
  marked as compatibility adapters and no new contract spelling reaches a source read without
  native admission.
- Confirmed all opaque checkpoint fields, including dedupe-window start/end bounds, are byte slices
  with defensive copying; JSON is used only for base64 byte preservation, never provider-token
  interpretation.
- Confirmed version-zero legacy cursor state turns into a typed invalid-checkpoint outcome before
  either warehouse or connector source reads.
- Confirmed generic destination writes do not advance state on errors or on a successful call that
  reports failed records; local warehouse raw, final, and materialized-dedupe files plus changed
  directories are synced before its acknowledgement path.
- Confirmed tombstone validation requires event identity, key/image shape, and ordered position;
  history exposes `_valid_from` and only a `_valid_to`/`_is_current=false` validity-window close
  as conforming.
- Confirmed native contract JSON has no REST API-surface/method/path or raw SQL/HTTP/shell field,
  and executable admission requires both a matching explicit interface and the embedded fixture
  evidence, including insert/update, both delete forms, invalidation, replay, and handoff cases.
- Ran `git diff --check`, focused unit tests, app/connector/CLI tests, `go vet`, build, and all
  non-suite `make verify` gates listed in `VERIFICATION.md`.

## Scope notes

- The old engine-derived five-mode projection and its public surfacing remain untouched because
  #3748/#3860 own migration/help/docs/website presentation. No new mode is advertised from this
  foundation.
- No #3745 descriptor or `commandrunner` function was changed; #3810 remains a reusable contract
  for their consumers rather than a duplicate descriptor/executor lane.
