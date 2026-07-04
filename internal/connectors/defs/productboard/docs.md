# Overview

Productboard is a wave2 fan-out declarative-HTTP migration. It reads Productboard features,
notes, components, and products through the Productboard public API
(`GET https://api.productboard.com/...`). This bundle targets capability parity with
`internal/connectors/productboard` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Productboard API access token via the `access_token` secret; it is sent as a standard
Bearer token (`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)`
(`productboard.go:157`) exactly. `base_url` defaults to `https://api.productboard.com`.

## Streams notes

All 4 streams read records from the `data` key and paginate via Productboard's own `links.next`
absolute-URL field (`pagination.type: next_url`, `next_url_path: "links.next"`), matching legacy's
own pagination loop exactly (`productboard.go:105-141`).

The first request sends `limit=100` (legacy's `defaultPageSize`) and, when `start_date` is
configured, `updated_since=<start_date>` (legacy: `query.Set("updated_since", start)`,
`productboard.go:114-116`) — modeled here via the opt-in optional-query dialect
(`"updated_since": {"template": "{{ config.start_date }}", "omit_when_absent": true}`), since
`start_date` is genuinely optional with no fallback default. **This is a one-shot, config-driven
filter, not a stateful incremental cursor**: legacy never persists or reads back a cursor value
from `req.State` for any of these 4 streams, so this bundle declares no `incremental` block for
any stream (declaring one would grant an `incremental_append` sync-mode capability legacy never
actually implements) — matching legacy's `CursorFields: []string{"updated_at"}` catalog metadata
on `features`/`notes` being purely cosmetic (never wired into any actual read-time filtering).

**`page` is deliberately NOT declared as a static per-stream query value**, unlike `limit`/
`updated_since`: the engine re-applies every `stream.Query` entry on EVERY page request, including
when following an absolute `next_url` (`read.go`'s `mergeQuery` + `resolveURL`'s `Del`-then-`Add`
query merge, which REPLACES rather than adds to a same-named param already present on the URL). A
static `page: "1"` would silently force every subsequent page's URL back to `page=1` — an actual
pagination-breaking bug (looping on page 1 forever), not a benign idempotent re-send like `limit`
or `updated_since` (whose values never change across pages for a single read). Productboard's
first page defaults to `page=1` when the param is omitted, so omitting it reproduces legacy's exact
first request while staying safe on every subsequent page, which follows the recorded absolute
`links.next` URL verbatim (already carrying Productboard's own correct page number).

Legacy's `mapRecord` (`productboard.go:170-180`) applies `name: first(item, "name", "title")`
shared across all 4 streams, and passes `status` through as a raw (frequently object-shaped, e.g.
`{"name": "Planned"}`) value with no flattening. Both are modeled via direct schema projection
(`name`, `title`, and `status` all declared as separate schema properties, matching legacy's raw
key-for-key emission exactly) rather than a `computed_fields` rename, since legacy itself emits
BOTH `name` (the fallback-resolved value) AND the raw `title` key side-by-side
(`"name": ..., "title": item["title"], ...`) — there is no rename to perform, only a
multi-candidate fallback for `name` specifically. The engine's `computed_fields` dialect has no
multi-key coalesce filter, so `name`'s exact 2-way fallback (`item["name"]` OR `item["title"]`) is
not modeled beyond direct projection of whichever raw key the record happens to carry; per
legacy's own fixture/fields (`productboard_test.go`'s `features` fixture sets `name` directly, its
second record omits `name` entirely relying on nothing — legacy's own test never actually exercises
the `title`-fallback branch either), this is not a proven divergence against any concretely
observed real record shape.

**Pagination fixtures are single-page**, per `conventions.md` §4's sanctioned `next_url` exception:
every stream's next-page URL is the replay server's own runtime-assigned address. A live 2-page
proof is out of scope for this wave (hard rule: no Go/paritytest packages) — see Known limits.

## Write actions & risks

None. Productboard's mutating endpoints are intentionally unsupported by legacy (package doc:
"avoids all mutating endpoints"); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **No live 2-page pagination proof for this wave.** All 4 streams use `next_url` pagination with
  single-page fixtures (the sanctioned `conventions.md` §4 exception); a live
  `paritytest/productboard`-style test is out of scope under this wave's hard rule prohibiting new
  Go/paritytest packages.
- **`start_date`/`updated_since` is a one-shot config filter, not a real incremental cursor.** No
  `incremental` block or `x-cursor-field` is declared for any stream, matching legacy's own lack of
  state-cursor read/persist logic exactly (see Streams notes above) — this is a faithful
  non-capability port, not a narrowing.
- **`name`'s exact 2-way raw-key fallback (`name` OR `title`) is approximated by direct projection
  of both keys side-by-side**, not a computed rename, since the engine's `computed_fields` dialect
  has no multi-key coalesce filter. Not exercised as a genuine divergence by legacy's own test
  fixtures (see Streams notes above).
- **Legacy's `raw` escape-hatch is preserved via `projection: "passthrough"`.** `mapRecord`
  stamps a full copy of the source item onto every record under the key `raw`
  (`productboard.go:178`), which is how every non-allowlisted real Productboard field
  (description, owner, timeframe, links, custom fields, etc.) reached the destination. All 4
  streams declare `projection: "passthrough"` so every raw source field survives unfiltered
  rather than being narrowed to the 6 schema-declared properties. Shape note: passthrough
  surfaces the source object's fields at the record top level, whereas legacy nested the full
  copy under a literal `raw` key — the substantive data (every real field) is preserved
  identically; only the nesting differs. The engine's `computed_fields` dialect has no
  whole-record reference primitive, so reproducing the exact literal `raw` nesting is not
  expressible, but passthrough eliminates the field-drop data-loss that a schema-mode allowlist
  would silently cause.
- **Legacy's fixture-mode-only synthetic records are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) emits synthetic records with an id shaped
  `"<stream>_<n>"` (not derived from any real API field); this bundle targets the live path only,
  matching every other wave1/wave2 bundle's documented precedent.
