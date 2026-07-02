# Overview

Gutendex is a wave2 fan-out declarative-HTTP migration. It reads Project Gutenberg books from the
free, public Gutendex JSON API (`GET https://gutendex.com/books/`) in four views (`books`,
`popular_books`, `latest_books`, `english_books`). This bundle targets capability parity with
`internal/connectors/gutendex` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

None. Gutendex is a public, unauthenticated API; this bundle declares no `auth` block at all
(matching legacy, which builds its `connsdk.Requester` with no `Auth` field set) and `spec.json`
has no `x-secret` field.

## Streams notes

All 4 streams read the single `/books/` resource with a different fixed query view, matching
legacy's `gutendexStreamEndpoints` table: `books` (no extra params, the API's default popular-first
order), `popular_books` (`sort=popular`), `latest_books` (`sort=descending`, highest Gutenberg ID
first), `english_books` (`languages=en`). Every stream also accepts the same optional user filters
legacy merges in (`search`, `languages`, `topic`, `sort`, `copyright`, `author_year_start`,
`author_year_end`, `ids`) via the opt-in optional-query dialect (`omit_when_absent: true`), left off
the request entirely when unset — matching legacy's `gutendexQuery`, which only sets a filter param
when the corresponding config value is a non-empty trimmed string. Per legacy's own precedence
rule ("stream params win over config so e.g. popular_books always sorts by popularity"), each
view's fixed param (`sort`/`languages`) is declared as a plain literal on that stream's `query` map
and is NOT also declared as a `config.*`-templated optional entry for that same key — so a
config-level `sort`/`languages` override can never contend with a view's own fixed value, matching
legacy's `out.Del(k)` overwrite exactly.

Pagination follows Gutendex's Django REST Framework convention (`pagination.type: next_url`,
`next_url_path: "next"`): the response's top-level `next` field is an absolute URL to the
following page (or `null` when exhausted), matching legacy's `harvest` loop exactly. Records live
at the `results` key.

Array-valued raw fields (`languages`, `subjects`, `bookshelves` — plain string arrays) are joined
into comma-separated strings via `computed_fields`' `join:,` filter, matching legacy's
`joinStrings` helper exactly (e.g. `"languages": "{{ record.languages | join:, }}"`).

No stream declares an `incremental` block: the Gutendex API has no updated-at field, matching
legacy's own comment ("the API has no updated-at field, so there is no incremental cursor
(full-refresh only)").

## Write actions & risks

None. Gutendex is read-only (`capabilities.write: false`); this bundle ships no `writes.json`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`authors`/`translators` (name-only, comma-joined) and the first-author-promoted
  `author_name`/`author_birth_year`/`author_death_year` fields are NOT modeled — `ENGINE_GAP`.**
  Legacy's `joinAuthorNames` extracts each author OBJECT's `name` sub-field before joining
  (`{"name": "Melville, Herman", ...}` → `"Melville, Herman"`), and `firstAuthor` reaches into
  the first element of the `authors` array to promote its `name`/`birth_year`/`death_year` onto
  dedicated top-level columns. The engine's `computed_fields` dialect has exactly two relevant
  primitives — a dotted `record.<path>` walk (map-key access only, `interpolate.go`'s
  `resolveRecordPathValue` type-asserts every path segment to `map[string]any` and hard-misses on
  a `[]any`, i.e. **no array-index addressing at all**) and `join:<sep>` (`applyJoinFilter`,
  which stringifies each array ELEMENT as a whole via `fmt.Sprint`, with no way to first project a
  sub-field out of each element) — neither can express "pluck `.name` out of every element of an
  array of objects, then join" or "read element `[0]`'s sub-field". This differs from searxng's
  `engines` precedent (ledger item 4/RESOLVED), which joins an array of plain STRINGS, not
  objects; gutendex's `authors`/`translators` are arrays of `{name, birth_year, death_year}`
  objects, a strictly harder shape the dialect does not yet cover. Faking this with a
  `computed_fields` template that only partially reproduces legacy's output (e.g. joining the raw
  objects' Go-map-print form) would silently diverge from legacy's actual comma-separated-names
  wire shape — an unacceptable deviation per conventions.md §5's meta-rule — so these 5 fields are
  dropped from this bundle's schemas entirely rather than approximated. This is a genuine
  `ENGINE_GAP` (a dialect gap recurring in a different shape from the array-of-strings case the
  engine already closed), not a Tier-2/Tier-3 escalation: everything else in this connector is
  fully Tier-1-expressible. A future engine increment adding an array "pluck a sub-field from every
  element" primitive (or a small `RecordHook`) would close this; until then, `id`, `title`,
  `subjects`, `bookshelves`, `languages`, `copyright`, `media_type`, and `download_count` are the
  full parity-correct field set this bundle emits.
- **Legacy's fixture-mode-only stamped fields (`stream`, `fixture`) are not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`) stamps two extra
  fixture-affordance fields onto every record that are not part of the live record shape. This
  bundle's schemas and fixtures target the live record shape only; the engine's own
  conformance/fixture-replay harness provides the credential-free test affordance legacy's fixture
  mode was built for.
- **`max_pages` (legacy's bounded-crawl default of 3 pages) is not modeled.** The engine's
  `next_url` paginator has no config-driven page-count knob (unlike `page_number`/`offset_limit`,
  which read `PaginationSpec.PageSize`); pagination is bounded only by the response's own `next`
  field going `null`, matching Gutendex's real termination signal. A caller wanting a bounded
  crawl of the 78k-book catalog must rely on the engine's read path, not a config value, to cap
  page count — the same gap class bitly documents for its own `next_url` stream.
