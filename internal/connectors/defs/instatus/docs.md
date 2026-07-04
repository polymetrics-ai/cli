# Overview

Instatus is a declarative HTTP connector for the authenticated Instatus REST API at `https://api.instatus.com`. It keeps the legacy read streams for status pages, components, incidents, and maintenances, then expands Pass B coverage to the documented status-page, workspace, team, template, notice, subscriber, metric, monitor, routing-rule, escalation-policy, and on-call schedule resources.

## Auth setup

Provide an Instatus API key via the `api_key` secret. The connector sends it as `Authorization: Bearer <api_key>`. `base_url` defaults to `https://api.instatus.com` and can be overridden for tests.

## Streams notes

The original legacy-parity streams keep their legacy record projection: `pages`, `components`, `incidents`, and `maintenances`. New Pass B streams use passthrough projection with permissive schemas so resource-specific API fields from the docs are preserved. Page-scoped streams require `page_id`; detail streams require the corresponding id config key, such as `component_id`, `incident_id`, `maintenance_id`, `metric_id`, or `audience_group_id`.

Most list endpoints use Instatus `page`/`per_page` pagination with a fixed page size of 50. Monitor list/log endpoints use the documented `page`/`limit` shape. The public status summary endpoint is intentionally excluded in `api_surface.json`: it is served from each public status-page host rather than the authenticated API base, and this bundle must not send the API bearer token to arbitrary public domains.

## Write actions & risks

`writes.json` covers dialect-expressible POST, PUT, and DELETE endpoints from the documented API surface. Delete actions are marked `confirm: destructive` and use idempotent 404 handling. Path identifiers are taken from each write record, while non-path fields are sent as the JSON body.

## Legacy parity

The legacy Go connector remains read-only and exposes only four streams. This bundle keeps those streams' record shape stable while adding documented Pass B streams and write actions for the declarative engine. Runtime-configurable legacy `page_size` and `max_pages` remain limited by the static pagination dialect; the bundle uses the legacy default page size of 50.

## Known limits

- The public status summary endpoint is documented as `your-status-page-url/summary.json`, so it is not represented as an authenticated API stream.
- Monitor run and monitor-group run GET endpoints are excluded because they trigger checks rather than return passive list/detail data.
