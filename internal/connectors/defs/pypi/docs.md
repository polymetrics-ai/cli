# Overview

PyPI is a wave2 fan-out declarative-HTTP migration, **partial**. It reads PyPI project metadata
through the PyPI JSON API (`GET https://pypi.org/pypi/<project>/json`). This bundle targets
capability parity with `internal/connectors/pypi` (the hand-written connector it migrates) for its
`project` stream only; the legacy package stays registered and unchanged until wave6's registry
flip, and remains the authoritative implementation for the `releases` stream (see Known limits).

## Auth setup

None. PyPI's JSON API is public and credential-free, matching legacy exactly (no `auth` config, no
secrets); `streams.json`'s `base.auth` declares `mode: none`. `base_url` defaults to
`https://pypi.org`.

## Streams notes

`project` reads `GET /pypi/<project_name>/json` and emits the response body's `info` object
**verbatim** (`records.path: "info"`, `projection: passthrough`), matching legacy's `emitProject`
exactly (`emit(connectors.Record(rec))` where `rec` is the raw `info` map — every field PyPI's JSON
API returns under `info` survives, not just legacy's narrower declared stream catalog of `name`/
`version`/`summary`). `project_name` is urlencoded into the path per-segment by the engine's default
path-interpolation filter, and PyPI's JSON API always returns the LATEST version's metadata for the
unversioned `/pypi/<project>/json` endpoint — matching legacy's own behavior when its `version`
config is unset (the default, and the only path this bundle implements — see Known limits).

## Write actions & risks

None. Legacy `pypi` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- **The `releases` stream is NOT implemented in this pass — ENGINE_GAP.** Legacy's `releases`
  stream has two config-dependent behaviors: (1) when `version` is set, it reads
  `pypi/<project>/<version>/json` and emits one record per file in that version's `urls[]` array
  (a flat array — would be expressible via `records.path: "urls"` plus `computed_fields` stamping
  the static `config.project_name`/`config.version` values); (2) when `version` is unset (legacy's
  own default, and what its own `TestReadReleasesMapsPyPIJSON` test exercises without setting a
  version), it reads the UNVERSIONED endpoint's `releases` object — a **map keyed by version
  string** (`{"1.0.0": [...], "2.0.0": [...]}`) — sorts the keys, flattens every key's file array
  into one record stream, and stamps **each record's `version` field with the key it came from**
  (`sort.Strings(versions)` + nested loop, `pypi.go:163-179`).

  The engine's `records.path` (`RecordsSpec`, `internal/connectors/engine/bundle.go`) can only
  select ONE array or single object at one static dotted path (`connsdk.RecordsAt`); it has no
  mechanism to iterate a JSON object's own keys, treat each key's value as a sub-array of records,
  and stamp the originating key onto each of that sub-array's records as a field. `computed_fields`
  cannot substitute either: it resolves against a single already-extracted record, never against
  "which map key produced this record" (there is no `record.<mapkey>`-style reflection primitive,
  and the value legacy needs — the sorted map key itself — is not present anywhere inside the file
  object being mapped). This is a genuine `records`-extraction dialect gap, not a workaround-able
  templating limitation: closing it needs either a `StreamHook` (Tier 2 — a custom `ReadStream`
  that performs the sort/flatten/stamp in Go) or a new engine `RecordsSpec` primitive (e.g. an
  "iterate object entries, stamp key as field X" mode) that does not exist today. Per this wave's
  hard rule against inventing Go for a declarative-only fan-out pass, `releases` is left unported;
  legacy's `internal/connectors/pypi` stays registered and authoritative for this stream. Also
  affected: the `version`-set sub-case, which IS individually expressible, was not implemented
  either, since a single `path` template cannot conditionally branch between
  `pypi/<project>/json` and `pypi/<project>/<version>/json` based on whether `version` resolves —
  the dialect's path templating always hard-errors on an absent referenced key, with no
  `omit_when_absent`-style tolerance for path segments (that tolerance exists only for `query`
  entries). Splitting this into two streams was considered and rejected: the no-version case (the
  one that actually needs its own stream to be useful) is exactly the inexpressible one, so
  splitting would not have produced a materially more complete bundle.
- **The `version` config key is not declared** in `spec.json` — since neither the `releases`
  stream nor a version-pinned `project` read path is implemented, a declared-but-unwireable
  `version` property would be dead config (conventions.md §2/F6).
- Full PyPI Warehouse API surface (the Simple API index, registry-wide `/stats`) is out of scope;
  see `api_surface.json`.
