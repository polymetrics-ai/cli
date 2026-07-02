# Overview

Insightful is a read-only declarative-HTTP migration (wave2 fan-out) of
`internal/connectors/insightful` (legacy Go package `insightful`). It reads Insightful
workforce-analytics employees, teams, projects, and directory entries through the Insightful REST
API (`https://app.insightful.io/api/v1`). This bundle targets capability parity with the legacy
connector; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Insightful API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged. `base_url` defaults to
`https://app.insightful.io/api/v1` and may be overridden for tests or proxies.

## Streams notes

Insightful's list endpoints return two different shapes, matching legacy's `extractRecords`
byte-sniffing exactly per stream (proven by legacy's own tests: `TestReadPaginatesAndAuthenticates`
exercises `/employee`'s enveloped shape, `TestReadTopLevelArray` exercises `/team`'s bare-array
shape):

- `employee`, `projects` (`GET /project`), `directory` are enveloped (`{"data": [...], "next":
  "<token>"}`) and paginated: `pagination.type: cursor` with `token_path: next` and `cursor_param:
  next`, matching legacy's `next`-token echo loop exactly (an absent/empty `next` in the response
  stops the read).
- `team` returns a bare top-level JSON array with no envelope and is not paginated in this bundle
  (`pagination.type: none`, `records.path: ""`), matching legacy's own observed behavior for this
  endpoint (legacy's generic cursor loop technically re-checks for a `next` token on every stream,
  but `/team`'s real response never carries one, so the loop always stops after one page in
  practice).

Primary key is `id` on every stream. `employee`/`projects`/`directory` declare `x-cursor-field:
updatedAt` and an `incremental` block (`request_param: start`), matching legacy's `start`
query-param behavior for the RESUMED-sync case: the app-persisted state cursor (the max `updatedAt`
across previously-emitted records, stored as its raw digits-only-millis string) is forwarded to
`start` **verbatim** (`param_format` omitted — the engine's default, no numeric reinterpretation),
which reproduces legacy's `incrementalLowerBound` exactly on every state-cursor-driven repeat sync.
`team` has no cursor field, matching legacy's `CursorFields` being unset for that stream.

## Write actions & risks

None. Insightful is read-only in this bundle (`capabilities.write: false`); legacy also rejects
every write with `connectors.ErrUnsupportedOperation`. No `writes.json` is shipped.

## Known limits

- Full Insightful API surface (time windows, screenshots, activity analytics, project/task writes)
  is out of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Only the 4 legacy-parity streams are
  implemented.
- **`start_date`-seeded FRESH syncs are not modeled** (documented scope narrowing, not a data-
  correctness change on the common resumed-sync path). Legacy's `incrementalLowerBound` computes
  the `start` query param two ways: (1) from the persisted state cursor — forwarded verbatim, fully
  reproduced above — or (2) on a brand-new sync with no prior cursor, by parsing the RFC3339
  `start_date` config value and converting it to Unix **milliseconds**
  (`t.UnixMilli()`). The engine's `param_format` dialect (`rfc3339`/`unix_seconds`/`date`/
  `github_date_range`) has no millisecond variant — `unix_seconds` would silently produce a value
  1000x too small (a real data-correctness bug, not cosmetic, since it would mis-scope the server-
  side filter), so this bundle deliberately does NOT declare `start_config_key` on any stream. The
  practical effect: a genuinely fresh sync (no persisted cursor yet) reads the full stream instead
  of legacy's `start_date`-bounded subset — strictly MORE data returned, never fewer or wrong
  records, so it does not violate the parity-deviation meta-rule (no accepted-input's emitted-
  record DATA is changed; only a scope-narrowing opt-in default is dropped). Once every subsequent
  sync resumes from a real persisted cursor, behavior is byte-for-byte identical to legacy. Adding
  a `unix_millis` `param_format` to the engine dialect would close this narrowing; filed as a
  candidate for a future engine mini-wave increment if this recurs on another connector (see
  `docs/migration/conventions.md` §6's `ENGINE_GAP` recurrence-threshold rule) rather than invented
  ad hoc here.
