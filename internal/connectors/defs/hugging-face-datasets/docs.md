# Overview

Hugging Face - Datasets is a wave2 fan-out declarative-HTTP migration, **partial**: it reads Hugging
Face dataset splits (`splits`) and per-split size metrics (`sizes`) through the public
dataset-viewer REST API (`GET https://datasets-server.huggingface.co/...`). The legacy connector's
3rd stream, `rows`, is NOT migrated here (see Known limits) and stays on
`internal/connectors/hugging-face-datasets` (the hand-written connector this bundle partially
migrates); that legacy package stays registered and unchanged until wave6's registry flip resolves
the blocker or a future wave adds a `RecordHook`.

## Auth setup

Authentication is OPTIONAL: public datasets are readable with no credentials at all, matching
legacy's own `apiToken`/`requester` behavior (`huggingfacedatasets.go:281-304`) exactly. Legacy checks
three possible secret keys in priority order (`api_token`, then `access_token`, then `token`,
first-non-empty wins) and sends whichever is found as a Bearer token. This bundle reproduces that
exact priority with three stacked `when`-gated auth candidates (the zendesk-support
dual-auth-ordering pattern, `docs/migration/conventions.md` §3, extended to three secrets):

```json
"auth": [
  { "mode": "bearer", "token": "{{ secrets.api_token }}", "when": "{{ secrets.api_token }}" },
  { "mode": "bearer", "token": "{{ secrets.access_token }}", "when": "{{ secrets.access_token }}" },
  { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ secrets.token }}" },
  { "mode": "none" }
]
```

— `api_token` wins when set (matching legacy's first-checked key), `access_token` is the fallback,
`token` is the last fallback, and `none` (no `Authorization` header at all) applies when none of the
three secrets are configured, exactly matching
`TestReadSplitsNoAuthHeaderWhenTokenAbsent`'s asserted no-header behavior. `base_url` defaults to
`https://datasets-server.huggingface.co` and may be overridden for tests/proxies.

## Streams notes

`dataset_name` (e.g. `ibm/duorc`) is a required config value sent as the `dataset` query param on
every request, matching legacy's `datasetName` resolution exactly.

`splits` (`GET /splits`) and `sizes` (`GET /size`) are both single-request, non-paginated endpoints
(`pagination: none`), matching legacy's `readList` behavior (`kindList` in
`huggingfacedatasets/streams.go`) exactly: `splits` records live at the top-level `splits` key;
`sizes` records live at the nested `size.splits` key (legacy's `recordsPath: "size.splits"`). Neither
stream declares an `incremental` block, matching legacy (the dataset-viewer API exposes no
incremental cursor field for either).

## Write actions & risks

None. The Hugging Face dataset-viewer API is read-only (legacy's own package doc: "It is read-only;
there is no reverse-ETL surface for this API"); `capabilities.write` is `false` and this bundle ships
no `writes.json`.

## Known limits

- **`rows` (the legacy `/rows` stream, offset-paginated dataset row reads) is NOT migrated — ENGINE_GAP
  blocker.** Legacy's `rowRecord` (`huggingfacedatasets/streams.go:114-130`) hoists every key of each
  row's nested `row` object to the TOP LEVEL of the emitted record (e.g. a dataset whose rows have a
  `text`/`label` schema emits top-level `text`/`label` fields, in addition to the nested `row` object
  itself) — the column set varies per dataset (it is whatever that dataset's own schema happens to
  contain), so it cannot be statically declared in a fixed JSON-schema `properties` map the way every
  other migrated stream's record shape can. `TestReadRowsPaginatesAndAuthenticates` asserts this
  flattening is load-bearing (`rec["text"] != nil` is checked directly on the emitted record, not just
  inside a nested `row` object), so silently dropping the top-level duplication and shipping ONLY the
  nested `row` object would be a real, observable data-shape change for any consumer that reads the
  legacy connector's top-level columns today — not a safe "just narrow the schema" scope reduction.
  The declarative dialect's `computed_fields` is a fixed named output-key map with no "spread an
  arbitrary-width `record.<path>` object's keys into the top level" primitive (unlike `join:<sep>` or
  `last_path_segment`, which operate on a single named field), so this cannot be expressed in Tier 1.
  Expressing it would need a `RecordHook` (`MapRecord`), which is out of this wave's JSON+docs.md-only
  scope (task hard rule 1) and would also make this bundle need 2 Tier-2 hook interfaces (`AuthHook`
  for the 3-secret fallback above, if that were ever converted to a hook too, plus `RecordHook` for
  this) — worth scoping as a dedicated Tier-2 `hooks/hugging-face-datasets/` package in a follow-up
  wave rather than a partial/approximate Tier-1 workaround here. `rows` stays on the legacy
  implementation in the interim; `bundles.streams_after` for this connector is 2, not the legacy
  count of 3.
- **Offset-based row pagination itself (the paginator shape `rows` would need) is otherwise a
  standard `offset_limit` fit** (`offset_param: offset`, `limit_param: length`) — the blocker above is
  specifically the per-record dynamic column hoist, not the pagination mechanics; a future
  `RecordHook`-equipped Tier-2 bundle for this connector should still use the engine's declarative
  `offset_limit` paginator for the request shape and add the hook only for the `MapRecord` flattening
  step.
