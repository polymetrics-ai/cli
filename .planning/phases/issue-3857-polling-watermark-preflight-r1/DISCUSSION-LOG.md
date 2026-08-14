# Discussion log — #3857 polling-watermark preflight

## Authority and settled inputs

- Issue #3857 is the scope authority. #3810 (merged through #3882) owns the
  seven canonical modes and durable checkpoint envelope; this slice consumes
  them without adding another mode vocabulary or a scalar cursor fallback.
- Issue #3856 (merged through #4078) owns the immutable v1 polling corpus and
  its no-skip runner. This slice validates against that corpus; it does not
  edit, replace, filter, or broaden its fixtures.
- #2986 and #3745–#3749 retain changefeed/CDC capability derivation. #3762
  retains bounded query taxonomy. Neither surface is changed by this PR.

## Chosen boundaries

1. The declaration is a native-database definition record in the optional,
   separate `polling_watermark.json` file. It contains only closed enums,
   catalog-discovered object selectors, bounded numeric limits, and registered
   executor references. It has no SQL text, HTTP path, shell fragment,
   caller-selected protocol, or fabricated REST command; it never belongs in
   `database.json` or `changefeed.json`.
2. `PollingPreflight` is a no-I/O runtime gate. It validates the declaration,
   resolves the exact registered source and apply executors, and checks their
   immutable-corpus registrations before exposing a resolved, immutable result.
   A later source/apply slice may consume that result; it cannot start a read
   from a merely syntactically valid declaration.
3. The preflight is intentionally distinct from the existing changefeed
   descriptor and from `commandrunner` REST preflight. Polling watermark is not
   promoted to change capture, and native database work has no `api_surface`
   row.
4. No native engine declaration, driver, database connection, credential, or
   destination DML is added. Tests use observable in-process fakes because
   #3857 explicitly forbids live database calls and #3858/#3859 own the real
   source/apply implementations.

## Inline GSD fallback

The generated `discuss-phase` and `plan-phase --tdd` prompts were resolved via
`scripts/gsd prompt`. This disposable non-Pi worker has no compatible isolated
Pi runtime, and the repository's single-worker contract forbids role spawning,
so the GSD discussion and plan are recorded inline in this phase directory.
