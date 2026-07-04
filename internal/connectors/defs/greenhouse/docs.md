# Overview

Greenhouse is a declarative HTTP connector for the official Greenhouse Harvest API. This Pass B bundle preserves the five legacy record projections for `candidates`, `applications`, `jobs`, `offers`, and `users`, then adds the remaining documented GET endpoints as passthrough streams plus non-deprecated mutation endpoints as write actions.

## Auth setup

Provide the Greenhouse Harvest API token as the `api_key` secret. It flows into HTTP Basic auth as the username with a blank password, matching legacy `connsdk.Basic(secret, "")`, and is never logged. `base_url` defaults to `https://harvest.greenhouse.io/v1` and can be overridden for tests or proxies.

`on_behalf_of_user_id` is optional and is sent as the `On-Behalf-Of` header when configured. Harvest documents this header as required for many audited mutations. The current declarative dialect has bundle-level headers rather than per-action headers, so the header is omitted unless explicitly supplied.

## Streams notes

The legacy streams remain first and keep schema projection so their emitted records match `internal/connectors/greenhouse`: `candidates`, `applications`, `jobs`, `offers`, and `users`. Newly added streams use `projection: passthrough` with permissive schemas because the legacy connector never shaped those records.

List-style streams send `per_page` from `config.page_size` and use Harvest's Link-header pagination. Detail and singleton streams disable pagination. Path-scoped streams require the matching config key named in `spec.json` when selected, for example `candidate_id`, `application_id`, `job_id`, or `field_type`; these keys are optional globally because unrelated streams do not need them.

## Write actions & risks

The bundle declares 57 write actions for documented POST, PUT, PATCH, and DELETE endpoints that the Tier-1 dialect can express as one HTTP request per record. Record schemas require path and query placeholders, then allow documented JSON body fields to pass through. Body-less deletes and query-only actions send no request body. Mutations that delete, reject, anonymize, merge, or remove data are marked `confirm: destructive`. Reverse ETL must still follow plan, preview, approval, and execute.

## Known limits

- Deprecated/deactivated Harvest v1 mutations, including v1 opening deletion and deprecated scheduled-interview create/update endpoints, are excluded as `deprecated` in `api_surface.json`.
- Newly added stream schemas are permissive passthrough schemas derived from documented response shapes, not hand-curated warehouse schemas. The existing five legacy streams remain narrow to preserve emitted-record parity.
- Write fixtures prove request-line and body construction for the declarative shape, but callers must provide the documented Greenhouse JSON body fields for the selected action.
- Every stream fixture ships a single page for Link-header pagination because fixture files cannot declare response `Link` headers. A single page with no `Link` header exercises the documented termination behavior.
- No incremental sync mode is derived for any stream, matching legacy's real unfiltered-every-sync behavior.
