# Overview

When I Work is a read-only declarative-HTTP connector migrated from
`internal/connectors/when-i-work` (legacy wave2 fan-out). It reads users, locations, positions, and
shifts from the When I Work REST API v2. This bundle is capability-parity with the legacy
hand-written connector; the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a When I Work account `email` and `password` (both secrets); they are sent as HTTP Basic
auth credentials on every request (`auth: [{"mode": "basic", "username": "{{ secrets.email }}",
"password": "{{ secrets.password }}"}]`) and are never logged. `base_url` defaults to
`https://api.wheniwork.com` and may be overridden for tests or proxies.

## Streams notes

4 streams: `users` (`GET /2/users`, records at `users`), `locations` (`GET /2/locations`, records
at `locations`), `positions` (`GET /2/positions`, records at `positions`), and `shifts` (`GET
/2/shifts`, records at `shifts`). Primary key is `["id"]` for all 4; none is incremental or
paginated — legacy's `Read` issues exactly one unparameterized request per stream (`r.Do(ctx,
http.MethodGet, endpoint.resource, nil, nil)`) and this bundle mirrors that exactly (no `query`
block, no `pagination` block).

`shifts.start_time`/`shifts.end_time` are declared as `string` in the schema (legacy typed them
`timestamp`, a legacy-only Field type with no draft-07 JSON Schema equivalent; the API returns them
as RFC3339-shaped strings).

## Write actions & risks

None. This bundle covers only the read surface legacy implemented; `capabilities.write` is `false`
and this bundle ships no `writes.json`.

## Known limits

- Only the 4 legacy-parity read streams are implemented; other When I Work endpoints (time-off
  requests, availability records, shift mutation, login/session) are out of scope for this
  migration wave — see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.
- No stream declares pagination: legacy issues exactly one request per stream with no paging loop,
  and this bundle mirrors that (no `pagination` block in `streams.json`).
