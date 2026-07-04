# Overview

Cisco Meraki reads and writes Meraki Dashboard API v1 organizations, and per-organization
networks, devices, admins, licenses, configuration templates, policy objects, branding policies,
SAML roles, and organization audit logs (configuration changes, API requests). This bundle
originally targeted full capability parity with the legacy hand-written connector
(`internal/connectors/cisco-meraki`, read-only), which stays registered and unchanged until wave6's
registry flip. This is a Pass B full-surface expansion (verified against the real Meraki OpenAPI
3.0.1 spec, 670 paths / 957 endpoint-methods total) scoped to the organization/network/device/admin
INVENTORY-AND-ADMINISTRATION surface this connector already targets — the vastly larger
per-network-product-type configuration surface (wireless, switch, appliance, camera, sensor,
systems-manager) is out of scope; see `api_surface.json`'s scope note and `Known limits` below.

## Auth setup

Provide a Meraki Dashboard API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.meraki.com/api/v1` and may be overridden for tests/proxies (validated as an absolute
http/https URL with a host).

## Streams notes

`organizations` (`GET /organizations`, top-level JSON array response, `records.path: ""`) lists
every organization the API key can access; primary key `id`.

`organization_networks`/`organization_devices`/`organization_admins`/`organization_licenses`/
`organization_config_templates`/`organization_policy_objects`/`organization_branding_policies`/
`organization_saml_roles`/`organization_configuration_changes`/`organization_api_requests` are all
**org-scoped**: legacy's own read path first lists every accessible organization, then reads the
per-organization sub-resource once per organization id, stamping `organizationId` onto every
emitted record; this bundle reproduces that exact pattern for all 10 org-scoped streams with the
engine's `stream.fan_out` dialect: `ids_from.request` issues a preliminary, fully-paginated
`GET /organizations?perPage={{ config.page_size }}` listing (the SAME endpoint the `organizations`
stream itself reads, with legacy's page-size query preserved, extracting `id` off each record);
`into.path_var` makes the resolved organization id referenceable in the stream's own `path` as
`{{ fanout.id }}` (e.g. `/organizations/{{ fanout.id }}/devices`);
`stamp_field: organizationId` writes the current organization id onto every emitted record of that
stream, after projection — exactly matching legacy's `harvest`'s manual `record["organizationId"]
= orgID` stamp for the original 3 fan-out streams, now extended to the 7 new Pass B streams.
Pagination, page size, and any future per-organization `max_pages` cap are independent per
organization (a fresh Link-header paginator per org), mirroring legacy's own per-org `harvest`
call.

All streams share the same pagination shape: Meraki's RFC 5988 `Link: <url>; rel="next"` header
convention (`pagination.type: link_header`), consumed by the engine's SSRF-guarded
`linkHeaderPaginator` (same-host-only by default). `perPage` is sent as `{{ config.page_size }}`
on the FIRST request of each sub-sequence only for the streams whose Meraki endpoint accepts a
`perPage` param (`organization_networks`/`organization_devices`/`organization_admins`/
`organization_licenses`/`organization_policy_objects`/`organization_configuration_changes`/
`organization_api_requests`); `organization_config_templates`/`organization_branding_policies`/
`organization_saml_roles` have no `perPage` query param declared because Meraki's own OpenAPI spec
documents none for these 3 (small, rarely-paginated org-tier lists) — the `link_header` paginator
still applies if a `Link` header is ever present. `page_size` defaults to 1000 (Meraki's own
page-size ceiling and legacy's `merakiDefaultPageSize`/`merakiMaxPageSize`).

None of the original 4 streams exposes a legacy-recognized incremental cursor field
(configuration/state snapshots, not event streams) — matching legacy's own catalog, which
publishes no `CursorFields` for any stream; all remain full-refresh only. Of the 7 new Pass B
streams, `organization_policy_objects` publishes `x-cursor-field: updatedAt` and
`organization_configuration_changes`/`organization_api_requests` publish `x-cursor-field: ts`
(matching the aha-style "publish the field, never request it as a server-side filter" pattern
already used elsewhere in this wave, since none of these endpoints' timestamp fields are wired as
an `incremental.request_param` here) — all 3 remain full-refresh reads regardless. Primary keys:
`id` for `organizations`/`organization_networks`/`organization_admins`/`organization_licenses`/
`organization_config_templates`/`organization_policy_objects`/`organization_saml_roles`; `serial`
for `organization_devices` (matching legacy's own choice of the hardware serial as the stable
device identifier); `[organizationId, name]` for `organization_branding_policies` (Meraki's own
OpenAPI response schema for this resource documents no `id` field at all, only `name`);
`[organizationId, ts, label]` for `organization_configuration_changes` and
`[organizationId, ts, path, method]` for `organization_api_requests` (both are audit-log resources
with no single-field natural key; a timestamp plus a disambiguating field is the closest available
composite).

`organization_licenses`' schema deliberately **omits** the raw API's `licenseKey` field (and the
nested `permanentlyQueuedLicenses[].licenseKey`) — a Meraki license key is a sensitive
entitlement/billing credential, not ordinary inventory metadata, and schema-mode projection
silently drops any undeclared field, so it never reaches an emitted record or a downstream
warehouse.

## Write actions & risks

`capabilities.write` is `true`. Thirteen actions cover the network/device/admin/configuration-template/
policy-object INVENTORY mutation surface: `create_network`/`update_network`/`delete_network`,
`update_device` (Meraki has no device-create endpoint; devices are claimed into inventory via the
excluded `/organizations/{organizationId}/claim` endpoint, not created via this connector),
`create_admin`/`update_admin`/`delete_admin`, `create_config_template`/`update_config_template`/
`delete_config_template`, and `create_policy_object`/`update_policy_object`/`delete_policy_object`.
All 4 delete-kind actions declare `delete.missing_ok_status: [404]` (idempotent delete) and carry
an irreversible-mutation risk note requiring approval; the remaining 9 creates/updates carry an
external-mutation risk note. Deliberately excluded from writes.json: anything security-sensitive
(SAML roles, SNMP credentials, login-security policy, org-wide SAML toggle) or org-cosmetic
(branding policies) — see `api_surface.json`'s `requires_elevated_scope`/`out_of_scope` entries.
This exceeds legacy's own read-only scope (`Write` unconditionally returned
`ErrUnsupportedOperation`) — a deliberate Pass B capability expansion, not a parity deviation,
since it only ADDS capability legacy never had.

## Known limits

- Legacy's config-overridable `max_pages` (a per-read hard page-count cap, defaulting to
  unlimited) has no engine-dialect equivalent: `PaginationSpec.MaxPages` is a static integer
  declared in `streams.json`, not a templated/config-driven value. This bundle declares no
  `max_pages` (absent = unbounded), matching legacy's own default (unlimited) behavior for every
  caller that never overrides it; a caller that previously set a numeric `max_pages` override has
  no equivalent knob here. `page_size` (which legacy also allows overriding) IS fully
  config-driven via `spec.json`'s `page_size` property (default 1000), since it flows through an
  ordinary per-request query parameter rather than a paginator-construction-time field.
- `fixtures/streams/organization_*/page_1.json` for every org-scoped stream ship a single fixture
  organization id's worth of preliminary `/organizations` listing data plus one page of the
  sub-resource itself, to exercise the fan-out path under conformance's replay harness; see
  `conformance`'s fixture-replay behavior for fan_out-declared streams.
- **The per-network-product-type configuration surface is entirely out of scope** (wireless
  SSIDs/RF profiles, switch ports/STP/QoS, appliance firewall/VPN/traffic-shaping, camera
  video/analytics settings, sensor readings, systems-manager device management — roughly 900 of the
  API's 957 total endpoint-methods): this connector targets organization/network/device/admin
  INVENTORY and administration, not per-product-type network configuration, which is a categorically
  different and vastly larger surface better served by a dedicated connector scoped to that domain.
  See `api_surface.json`'s scope note.
- Security-sensitive org-tier settings (SNMP credentials, login-security policy, SAML SSO
  enablement/roles) are excluded from both streams and writes: several of these endpoints return or
  accept credential-shaped values (SNMP community strings, SAML role access-grants) this connector's
  dialect deliberately keeps out of ordinary record-shaped read/write bodies.
