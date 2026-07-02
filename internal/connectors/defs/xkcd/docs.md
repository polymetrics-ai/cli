# Overview

XKCD is the wave1-pilot Tier-1 declarative-HTTP migration (P-1, `.planning/phases/wave1-pilot/`)
of `internal/connectors/xkcd` (186 loc). It reads public XKCD comic metadata from the JSON API
(`GET {base_url}/info.0.json` for the latest comic, `GET {base_url}/{comic_number}/info.0.json`
for a specific comic). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/xkcd` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until the wave6 registry flip.

## Auth setup

No credentials are required: the public XKCD JSON API has no authentication. `base_url` defaults
to `https://xkcd.com` (matching legacy's `defaultBaseURL`, xkcd.go:18) but, unlike legacy, the
engine has no runtime default-value substitution for an unconfigured `config.*` key — every caller
must supply `base_url` explicitly (see "Known limits").

## Streams notes

Both streams return a **single JSON object**, not an array, so `records.single_object: true` with
`records.path: "."` (the whole body is the record) — matching legacy's direct
`json.Unmarshal(resp.Body, &rec)` (xkcd.go:93-97). No pagination, no incremental: XKCD's JSON API
has no list endpoint and no cursor field to page or filter by, matching legacy's own lack of any
pagination/incremental logic.

- **`latest`** (`GET /info.0.json`): the current/latest comic.
- **`comic`** (`GET /{{ config.comic_number }}/info.0.json`): a specific comic selected by the
  `comic_number` config value — a **templated path segment**
  (wave0's F1 stream-path-interpolation fix's first pilot consumer). The engine's
  `InterpolatePath` urlencodes the resolved segment by default and rejects a resolved value that is
  (or percent-decodes to) a bare `".."` segment, matching legacy's own guard (xkcd.go:83-87:
  `comic_number` must be non-empty and must not contain `/?#`). A hostile `comic_number` (e.g.
  `../../etc/passwd`) fails closed on BOTH sides before any request is ever issued —
  `TestParityXkcd_HostileComicNumberFailsClosedOnBothSides` in
  `internal/connectors/paritytest/xkcd/parity_test.go` asserts zero requests reach the upstream
  server on either connector.

Schema fields (`num`, `title`, `safe_title`, `year`, `month`, `day`) are copied field-for-field from
legacy's `Catalog()` field list (xkcd.go:55) and its direct pass-through record shape; `num` is the
primary key on both streams.

## Write actions & risks

None. XKCD is a read-only comic-metadata API with no mutation endpoints; `capabilities.write` is
`false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation` unconditionally (xkcd.go:108-110).

## Known limits

- **`base_url` has no engine-level default.** Legacy falls back to `https://xkcd.com` when
  `config.base_url` is unset (xkcd.go:124-128, `baseURL` helper). The engine dialect has no
  default-value substitution mechanism for an absent `config.*` reference in general `Interpolate`/
  `InterpolatePath` resolution (only `auth`'s `when` grammar tolerates absent-key-falsy; see
  `docs/migration/conventions.md` §3) — an unconfigured `base_url` is a hard error naming the
  missing key, not a silent fallback to the legacy default. This is a documented, ACCEPTABLE
  parity deviation (conventions.md §5 meta-rule): the only affected input is "no `base_url`
  configured at all", which is a genuinely different (stricter, fail-loud) behavior than legacy's
  silent default — a caller MUST configure `base_url` explicitly for this bundle. Every parity test
  in `internal/connectors/paritytest/xkcd/parity_test.go` supplies `base_url` explicitly on both
  sides, so this does not affect any parity assertion; it is called out here for anyone
  provisioning a production connection. `spec.json`'s `base_url` still declares
  `"default": "https://xkcd.com"` as a documentation/CLI-affordance annotation (matching stripe's
  own `base_url` pattern), even though the engine itself does not consume that annotation at
  runtime.
- **Fixture mode is a legacy-only affordance, not part of this bundle.** Legacy's `mode: fixture`
  config value short-circuits all network access and emits a synthetic record carrying `fixture:
  true` and a derived `stream` marker field (xkcd.go:100-106, `readFixture`) — this is a
  test/conformance-harness affordance on the legacy side only. Parity is asserted against legacy's
  LIVE read path via httptest (SPEC.md §5.1), which never emits `fixture` or `stream`; this bundle
  correspondingly does not stamp a `stream` computed_field (an earlier draft of this bundle
  incorrectly did — corrected before parity tests went green, since the live path never carries
  that field and adding it would have been a genuine record-shape deviation, not an acceptable one).
