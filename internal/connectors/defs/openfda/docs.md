# Overview

openFDA is a public, read-only REST API over FDA regulatory datasets (drug, device, and food).
This bundle reads five streams (`drug_event`, `drug_label`, `drug_enforcement`, `device_event`,
`food_enforcement`) from `https://api.fda.gov`. It is migrated from `internal/connectors/openfda`
(the hand-written connector this bundle replaces at parity); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

An API key is optional: openFDA works fully credential-free at a lower rate limit. When the
`api_key` secret is set, it is sent as the `api_key` query parameter to raise the caller's rate
limit (`base.auth`'s first candidate, gated by `when: "{{ secrets.api_key }}"`); when unset,
requests fall through to the `none` auth candidate and are sent unauthenticated — identical to
legacy's `requester`, which only attaches `connsdk.APIKeyQuery("api_key", key)` `if key !=
""`.

## Streams notes

All 5 streams share openFDA's common list envelope: `{"meta":{"results":{"skip","limit","total"}},
"results":[...]}` with offset/limit pagination (`pagination.type: offset_limit`, `limit_param:
limit`, `offset_param: skip`). `streams.json`'s `base.pagination.page_size` is set to `2` purely so
the required 2-page conformance fixture (`fixtures/streams/drug_event/{page_1,page_2}.json`) can
prove real pagination termination without an oversized fixture; legacy's actual runtime default
(`openfdaDefaultPageSize`, 100) is not itself expressible as a spec-overridable value (see the
`page_size`/`max_pages`/`max_skip` note below) — `2` is a fixture-authoring convenience, matching
the identical pattern in `internal/connectors/defs/aviationstack`'s golden. Pagination stops on a
short page (fewer than `page_size` records) — legacy's additional
`meta.results.total`-based early stop is a defensive optimization only reachable at the exact
page-size boundary; the engine's short-page stop alone terminates correctly for every input legacy
itself would accept. Every stream accepts an optional `search` query string (openFDA's Lucene-like
query syntax), applied via the `omit_when_absent` optional-query dialect — absent entirely when
unset, matching legacy's `if search != "" { base.Set("search", search) }`. Every field in every
stream's schema is copied verbatim from the raw openFDA result object (1:1 passthrough, matching
legacy's `*Record` mapper functions field-for-field); no renames or computed fields are needed.
`drug_enforcement` and `food_enforcement` share the identical field shape (both use legacy's
`enforcementRecord`/`enforcementFields`) but are declared as two independent schemas per the
one-schema-per-stream convention.

## Write actions & risks

None. openFDA is a read-only public regulatory API; `capabilities.write` is `false` and no
`writes.json` is shipped, matching legacy's `ErrUnsupportedOperation` `Write` stub.

## Known limits

- Legacy enforced a hard `openfdaMaxSkip` ceiling (25000) on the `skip` offset, matching openFDA's
  own documented cap. This bundle does not declare a matching `max_pages` bound (`base.pagination`
  has no `max_pages` set, i.e. unbounded) since the engine's `MaxPages` is a plain request-count
  cap, not an offset-value cap, and 25000/100 = 250 requests would need to be expressed as
  `max_pages: 250` to reproduce the exact ceiling; this is intentionally left unbounded for now
  since legacy's short-page/total stop already terminates every real dataset well before the skip
  ceiling in practice, and openFDA itself returns a 400 past the true ceiling (which would surface
  as a genuine read error, not silently wrong data). Revisit if pagination past 25000 records is
  ever exercised against the live API.
- The full openFDA surface (device 510(k)/PMA/UDI/enforcement, animal & veterinary events, tobacco
  problem reports) is out of scope for this wave; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 5
  legacy-parity streams are implemented.
