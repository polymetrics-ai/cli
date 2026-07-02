# Overview

Smaily is a wave2 fan-out declarative-HTTP migration. It reads Smaily campaigns, segments,
contacts, templates, and automations through the Smaily PHP API (`GET
https://<subdomain>.sendsmaily.net/api/*.php`). This bundle migrates
`internal/connectors/smaily` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide `api_username` (config) and `api_password` (secret); they are sent as HTTP Basic auth
credentials (`basic` auth mode), matching legacy's `connsdk.Basic(user, pass)` (`smaily.go:144`).
`base_url` is required directly by this bundle (see Known limits for why legacy's
`api_subdomain`-derived default is not modeled).

## Streams notes

All five streams (`campaigns`, `segments`, `subscribers`, `templates`, `automations`) hit their
own `GET api/<resource>.php` endpoint, matching legacy's `streamEndpoints` map exactly
(`campaign.php`, `segment.php`, `contact.php`, `template.php`, `autoresponder.php`). Records are
extracted from the response body's top-level array (`records.path: ""`, root-array shape),
matching legacy's `recordsPath: ""` for every stream and `connsdk.RecordsAt`'s empty-path ==
body-root semantics. None of the streams paginate in legacy (a single `r.Do` call per read, no
loop, and no query parameters at all) — `pagination.type: none` is declared, one request per
read, no query. Records carry `id` (integer primary key), `name`, and `created_at`, the exact
field set legacy's `streams()` catalog declares for every stream.

## Write actions & risks

None. Smaily's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`api_subdomain`-derived `base_url` is not modeled; `base_url` is required directly instead.**
  Legacy derives `https://<api_subdomain>.sendsmaily.net` from a separate `api_subdomain` config
  value when `base_url` is unset (`smaily.go:151-158`), including a subdomain-label safety check
  (`strings.ContainsAny(subdomain, "/:@")`). The engine's `spec.json` `"default"` materialization
  mechanism (`docs/migration/conventions.md` §3) only fills a FIXED literal default, not a
  value derived from another config key at read time — there is no declarative base-URL-
  construction template in this dialect. Per convention, this bundle narrows the config surface:
  `base_url` is a required spec property with no default, and `api_subdomain` is not declared at
  all (a declared-but-unwireable config key is worse than an absent one, per the searxng/bitly
  precedent). An operator who previously configured only `api_subdomain` must now supply the full
  `https://<subdomain>.sendsmaily.net` URL as `base_url`; this is a documented config-surface
  narrowing, not a data-shape change — every record emitted for a given account is identical
  either way once `base_url` resolves to the same origin.
