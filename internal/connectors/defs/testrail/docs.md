# Overview

TestRail is a wave2 fan-out declarative-HTTP migration of `internal/connectors/testrail` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads TestRail projects through the TestRail v2 REST
API (`GET <base_url>/index.php?/api/v2/get_projects`, TestRail's PHP-front-controller URL
convention). Read-only.

## Auth setup

Provide a TestRail username (email) via the `username` config value and a password or API key via
the `password` secret; both are required. They are sent as HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's
`connsdk.Basic(username, password)` (`testrail.go:101`). `password` is never logged. `base_url`
defaults to `https://example.testrail.io` (legacy's own placeholder default) and should be
overridden to the operator's real TestRail instance URL.

## Streams notes

`projects` is the only stream: `GET index.php?/api/v2/get_projects`, no pagination (a single
request; legacy issues exactly one `r.Do` call with no page loop, `testrail.go:71-84`), records at
the response body's array root (`records.path: "."`), primary key `["id"]`. TestRail's front
controller URL convention embeds the real API path as a raw (unencoded, no `=`) query string
segment rather than a normal path segment — `index.php?/api/v2/get_projects` is a STATIC LITERAL
(no `{{ }}` templating) in both `stream.path` and `base.check.path`, so it passes through
`InterpolatePath` completely unmodified (the urlencode-by-default filter only ever touches
resolved `{{ }}` template values, never literal surrounding text) and `url.Parse` then correctly
splits it into `Path: /index.php` + `RawQuery: /api/v2/get_projects` exactly as legacy's own
`connsdk.Requester` does — no dot-dot traversal risk (the literal contains no `..` segment).

`id` is a JSON integer, `name`/`announcement` are nullable strings (TestRail omits `announcement`
on some projects — legacy's own test fixture shows a project with no `announcement` key at all),
`is_completed` is a nullable boolean. No field renames are needed; plain schema projection copies
every field by exact key match.

## Write actions & risks

None. Legacy `testrail` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full TestRail API surface (test cases, test runs, test results, suites, milestones) is out of
  scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the single legacy-parity `projects` stream is implemented.
- No pagination is modeled, matching legacy exactly: `get_projects` is fetched in one request with
  no page/cursor loop of any kind in the legacy implementation.
- The `fixtures/streams/projects/page_1.json` fixture records TestRail's PHP-front-controller
  request shape as the replay harness sees it after Go's own `url.Parse` split
  (`path: "/index.php"`, `query: {"/api/v2/get_projects": ""}`) rather than the pre-split literal
  string, since fixture-request matching compares against the parsed `*http.Request`'s
  `URL.Path`/`URL.Query()`, not the original unparsed path template.
