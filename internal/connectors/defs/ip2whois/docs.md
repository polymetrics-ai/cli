# Overview

IP2WHOIS is a read-only WHOIS-lookup source connector. It exposes a single domain-lookup endpoint
(`GET https://api.ip2whois.com/v2?key=<key>&domain=<domain>`) that returns one WHOIS object per
domain, rather than a paginated list. This bundle migrates `internal/connectors/ip2whois` (the
hand-written legacy connector); the legacy package stays registered and unchanged until wave6's
registry flip. **This is a partial migration** — the `whois` and 4 per-role contact streams are
fully ported; the `nameservers` stream is blocked on a genuine engine gap (see Known limits).

## Auth setup

Provide an IP2WHOIS API key via the `api_key` secret; it is sent as the `key` query parameter on
every lookup (`{"mode": "api_key_query", "param": "key", "value": "{{ secrets.api_key }}"}`),
matching legacy's `connsdk.APIKeyQuery("key", secret)` wiring exactly. It is never logged.

## Streams notes

Legacy iterates a config-supplied, comma-separated domain list, performing one lookup per domain
and fanning each single-object response out into per-stream records. This bundle expresses the
per-domain iteration with the `fan_out` dialect (`fan_out.ids_from.config_key: "domains"`,
`fan_out.into.query_param: "domain"`) on every stream — the engine resolves the `domains` config
value once, splits/trims/drops empty entries, and repeats each stream's full
request/records/computed_fields sequence once per domain, exactly matching legacy's
`harvest`/`readFixture` per-domain loop. Legacy's separate single-value `domain` config field is
not modeled; a single domain is simply a length-1 `domains` list (documented config-surface
narrowing — see Known limits).

- **`whois`** (`records.path: ""`, one record per domain — the whole lookup object, no
  `stamp_field` needed since `domain` is already a field on that object): `computed_fields`
  flattens the nested `registrar`/`registrant`/`admin`/`tech`/`billing` sub-objects into
  `registrar_name`, `registrant_email`, etc. (dotted `record.<path>` references walk nested JSON
  directly — `{{ record.registrar.name }}`), and joins the raw `nameservers[]` array into a
  comma-separated string via the `join:,` filter, matching legacy's `whoisRecord`/
  `joinNameservers` output exactly. `update_date` is schema-declared as `x-cursor-field` (legacy
  publishes it in `CursorFields`) but no `incremental` block is declared — legacy never wires a
  request-time filter for it either (per `conventions.md` §8 rule 2: CursorFields-without-a-real-
  server-filter keeps `x-cursor-field` in the schema only).
- **`contacts_registrant`/`contacts_admin`/`contacts_tech`/`contacts_billing`** (one stream per
  legacy contact role, each `records.path` pointed directly at that role's nested sub-object —
  `registrant`/`admin`/`tech`/`billing`): each yields exactly 0 or 1 record per domain (present iff
  the sub-object is present on that domain's lookup response — `RecordsAt`'s ordinary
  single-object-at-path behavior naturally handles the 0-or-1 case, matching legacy's
  `contactRecords`' per-role `len(contact) == 0` skip). `fan_out.stamp_field: "domain"` stamps the
  domain to each contact record's `domain` field (unlike `whois`, the sub-object itself carries no
  `domain` field once `records.path` narrows into it); a static-literal `computed_fields.role`
  entry stamps the fixed role name (`"registrant"`/`"admin"`/`"tech"`/`"billing"`), matching
  legacy's `contactRecords` per-role loop. **Deviation from legacy**: legacy publishes these 4 roles
  under ONE catalog stream named `contacts` (primary key `[domain, role]`); this bundle publishes 4
  separately named streams instead, because `records.path` supports exactly one dotted path per
  stream declaration — there is no declarative way to select several independent named sub-object
  keys into a single stream. Every field legacy emits for every role is preserved verbatim (name,
  organization, street_address, city, region, zip_code, country, phone, fax, email, plus
  domain/role); only the catalog stream-name/count changed, not any record's data for any domain a
  legacy-valid config would produce — see the parity-deviation ledger.
- **`nameservers` is NOT implemented — see Known limits.**

`check` issues a single bounded lookup of the configured `domains` value, mirroring legacy's
`Check` implementation (a single-domain probe confirms auth and connectivity without mutating
anything).

## Write actions & risks

None. IP2WHOIS is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation` and there is no reverse-ETL write target for this API.

## Known limits

- **Blocked: `nameservers` stream (`ENGINE_GAP`).** Legacy's `nameserverRecords` fans the raw
  `nameservers` field — a bare JSON array of scalar strings (`["ns1.example.com",
  "ns2.example.com"]`), not an array of objects or an object-keyed map — out into one record per
  nameserver (`{domain, nameserver}`). The engine's only two array/object record-extraction
  primitives are `connsdk.RecordsAt` (an array is walked, but only elements that decode as a JSON
  OBJECT are kept — `internal/connectors/connsdk/extract.go:43-49` — a bare string element is
  silently dropped, yielding zero records) and `records.keyed_object` (explodes a JSON OBJECT's
  VALUES, which themselves must be objects — `internal/connectors/engine/read.go`'s
  `recordsAtKeyed` skips any non-object value the same way). Neither primitive fans a scalar-valued
  array into one record per element. Emitting a single joined string (the `whois` stream's
  `nameservers` field already does this via `join:,`) would change record CARDINALITY versus
  legacy's genuine 1-record-per-nameserver stream for any domain with 1+ nameservers — an
  accepted-input emitted-DATA change, not a cosmetic deviation, so it fails the meta-rule and is
  correctly left as a gap rather than silently approximated. This is a narrower, different gap than
  this connector's original quarantine reason (per-domain request iteration + whole-object fan-out,
  both now solved by the `fan_out` dialect) — see `docs/migration/quarantine.json`'s prior entry for
  the superseded reasoning.
- Legacy's separate single-value `domain` config field is not modeled; only the comma-separated
  `domains` list is (a single domain is a length-1 list). Declared-but-unwireable convenience
  aliases are not carried forward (F6, `conventions.md`).
- Only the `whois`/`contacts_*` streams are implemented; `nameservers` keeps its legacy
  implementation (`internal/connectors/ip2whois`) until the engine gap above is closed.
