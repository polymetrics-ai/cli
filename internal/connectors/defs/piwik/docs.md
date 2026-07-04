# Overview

Piwik / Matomo reads from Matomo's Reporting API through the single `GET /index.php?module=API` entry point. The bundle keeps the four legacy streams byte-for-byte compatible at the record boundary (`sites`, `visits`, `actions`, and `goals`) and expands Pass B coverage to Matomo's documented analytics report method inventory from `API.getReportMetadata`, plus selected site list/detail reads and live counters documented in the Reporting API guide.

## Auth setup

Provide `token_auth` as a secret; it is sent as the `token_auth` query parameter on every request and never logged. `base_url` defaults to the same placeholder legacy used (`https://matomo.example.com`), so live use must provide the real Matomo instance origin. `site_id`, `period`, and `date` scope analytics report streams; `custom_dimension_id` is only needed for `custom_dimensions_custom_dimension`, and `last_minutes` controls `live_counters`.

## Streams notes

The legacy streams preserve the hand-written connector's record mapping: `sites` renames `idsite` to `site_id`, `visits` renames visit identifiers and `lastActionDateTime`, `actions` renames `nb_hits`/`nb_visits` to `hits`/`visits`, and `goals` renames `idgoal` to `goal_id`. New analytics streams use the documented `Module.action` method name as the Matomo `method` query value and project documented metric fields with either a report-level primary key for summary objects or a computed `record_id` from Matomo's row `label` for dimensioned reports.

Matomo's Reporting API documents `filter_limit` and `filter_offset` as standard optional filters for report tables, so table and summary report streams use the engine's offset/limit pagination with a page size of 100. Summary endpoints naturally stop after their single object response; table endpoints stop on a short page.

## Write actions & risks

None. The legacy connector's `Write` returns `connectors.ErrUnsupportedOperation`, and Matomo's documented website/user/goal management mutations are GET-style RPC methods rather than POST/PUT/PATCH/DELETE endpoints. This bundle remains `capabilities.write: false` and ships no `writes.json`.

## Known limits

Matomo's full HTTP API is partly per-installation and plugin-dependent. This bundle covers the analytics report surface exposed by the official public `API.getReportMetadata` endpoint and explicitly enumerates common documented generic, management, binary, and helper methods in `api_surface.json` with typed exclusions. Premium/plugin reports that appear in a Matomo instance's metadata list but are not installed on another instance may return permission or plugin errors at runtime; they are still modeled because the public documentation exposes them as report methods.

The legacy `id_site` alias is still narrowed to `site_id` because the declarative query dialect has no coalesce operator. The legacy `page_size`/`max_pages` runtime overrides also remain absent; pagination sizes are bundle-fixed by the engine dialect. The legacy paginated streams (`visits`, `actions`) keep the legacy default page cap with static `max_pages: 3`.
