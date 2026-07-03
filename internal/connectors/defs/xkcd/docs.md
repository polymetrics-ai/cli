# Overview

XKCD is the wave1-pilot Tier-1 declarative-HTTP migration (P-1, `.planning/phases/wave1-pilot/`)
of `internal/connectors/xkcd` (186 loc). It reads public XKCD comic metadata from the JSON API
(`GET {base_url}/info.0.json` for the latest comic, `GET {base_url}/{comic_number}/info.0.json`
for a specific comic). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/xkcd` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until the wave6 registry flip.

**Pass B full-surface expansion (2026-07-03)**: re-verified against the live docs page
(`https://xkcd.com/json.html`) and live JSON responses (`/info.0.json`, `/614/info.0.json`) —
xkcd's entire documented JSON interface is exactly the two endpoint shapes already implemented
as the `latest`/`comic` streams below. There is no additional list/search/index JSON endpoint to
add as a stream (the JSON API has no comic-enumeration capability at all — a caller must already
know a comic number to fetch it) and no mutation endpoint of any kind to add as a
`writes.json` action (xkcd is a static, unauthenticated, publicly-cached site: no accounts, no
user-owned data, no documented write path — nothing to express, not a scope cut). `api_surface.json`
was updated to record this closed-surface finding explicitly rather than leaving a generic "Pass B
capability expansion" placeholder; every endpoint is `covered_by` a stream (zero `excluded`
entries needed, since there is no undocumented/skipped endpoint left over). No new streams, no
`writes.json`, no hook additions were needed — the bundle was already at full documented-surface
coverage.

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

Both streams declare `"projection": "passthrough"` (`engine/read.go`'s `projectRecord`), matching
legacy's own record shape exactly: legacy's read path is a raw passthrough
(`json.Unmarshal(resp.Body, &rec); emit(rec)`, xkcd.go:93-97) that emits every field of whatever the
real XKCD API returns, not a fixed subset. In `"schema"` projection mode (the engine default),
only schema-declared properties survive a read; an earlier draft of this bundle used that default
and a 6-field schema copied from legacy's `Catalog()` field list (xkcd.go:55) — Catalog() is a
capability-discovery listing, NOT legacy's record-shaping function, and does not describe every
field the live read path actually emits. That drift silently dropped `link`, `news`, `transcript`,
`alt`, and `img` from every real response (REVIEW-B.md finding 1, BLOCKER). Fixed by switching both
streams to `passthrough` — the schema's 11 declared `properties` (`num`, `title`, `safe_title`,
`year`, `month`, `day`, `link`, `news`, `transcript`, `alt`, `img`, matching the real XKCD JSON
API's documented field list, https://xkcd.com/json.html) now serve purely as a documentation/
validation surface (`required`, `x-primary-key`, types) — the actual emitted record shape is
whatever the live API returns, unfiltered, exactly like legacy. `num` is the primary key on both
streams.

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
  provisioning a production connection. `spec.json`'s `base_url` property does NOT declare a
  `"default"` key (no CLI-affordance annotation, unlike stripe's own `base_url` pattern) — its
  `description` documents the legacy default value in prose only, which is the sole place this
  default is recorded in this bundle (corrected: a prior revision of this doc claimed `spec.json`
  declared a `"default"` key here — it never has; REVIEW-B.md finding 2).
- **Fixture mode is a legacy-only affordance, not part of this bundle.** Legacy's `mode: fixture`
  config value short-circuits all network access and emits a synthetic record carrying `fixture:
  true` and a derived `stream` marker field (xkcd.go:100-106, `readFixture`) — this is a
  test/conformance-harness affordance on the legacy side only. Parity is asserted against legacy's
  LIVE read path via httptest (SPEC.md §5.1), which never emits `fixture` or `stream`; this bundle
  correspondingly does not stamp a `stream` computed_field (an earlier draft of this bundle
  incorrectly did — corrected before parity tests went green, since the live path never carries
  that field and adding it would have been a genuine record-shape deviation, not an acceptable one).
