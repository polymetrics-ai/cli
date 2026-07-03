# Overview

TallyPrime is a Tier-3 native source connector (design §B.7) built directly against
`docs/migration/quarantine.json`'s NON_REST finding: "TallyPrime = local TDL envelope gateway
(XML and JSON modes both envelope/RPC, POST to localhost:9000, no REST resources/pagination) —
Tier-3 native build queued". There is no legacy `internal/connectors/tally-prime` package to port
from parity — this connector is a new build, not a migration. TallyPrime exposes exactly one wire
endpoint (its local Gateway Server, default `http://localhost:9000`); every logical object it
understands (companies, ledgers, groups, stock items, vouchers) is requested by POSTing a different
TDL Export/Collection ENVELOPE/HEADER/BODY document to that same endpoint, never a REST
resource/path — the connector cannot be expressed as the engine's declarative HTTP dialect
(no per-stream `path`, no REST pagination) for the same reasons `native/postgres` and
`native/amazon-sqs` are Tier-3: an RPC/envelope protocol with no declarative equivalent. This
package follows `internal/connectors/native/postgres/`'s component split as the golden pattern
(`connector.go`/`connection.go`/`reader.go`/`cataloger.go`, each under 400 lines).

## Auth setup

TallyPrime's Gateway Server has no token/credential-based authentication of its own — it is a
loopback/LAN-only local service that trusts any process able to reach its configured port (this is
TallyPrime's own security model, not a gap in this connector). Configuration is therefore limited
to addressing and scoping, not secrets:

- `gateway_url` (config, default `http://localhost:9000`) — the base URL of the locally-running
  TallyPrime Gateway Server. TallyPrime must have "Act as a TDL/Gateway Server" enabled (Tally's
  own `F1: Help > TDL & Add-On > Set Component Configuration`/Gateway Server settings) and be
  listening on this address before Check/Read can succeed.
- `company` (config, required) — the exact TallyPrime company name every request is scoped to,
  sent as `STATICVARIABLES.SVCURRENTCOMPANY`. TallyPrime resolves every master/voucher relative to
  whichever company is active in the request context.
- `envelope_format` (config, default `json`) — `json` selects TallyPrime 7.0+'s native JSON export
  mode (`SVEXPORTFORMAT` = `$$SysName:UTF8JSON`); `xml` is the universally-supported fallback
  (`SVEXPORTFORMAT` = `$$SysName:XML`) for TallyPrime releases that predate native JSON export.
  There are no API keys, bearer tokens, or OAuth flows anywhere in this connector; `x-secret`
  fields do not appear in `spec.json` because none of TallyPrime's Gateway Server protocol carries
  a secret.

## Streams notes

Every stream is read by building one TDL Export/Collection envelope, POSTing it to `gateway_url`,
and decoding the response envelope's `BODY.DATA` (JSON mode) or `BODY.DESC.EXPORT` records (XML
mode) into `connectors.Record`s. The request envelope shape (both encodings share the same
skeleton, only `STATICVARIABLES.SVEXPORTFORMAT` and the wire body content-type differ):

```xml
<ENVELOPE>
  <HEADER>
    <VERSION>1</VERSION>
    <TALLYREQUEST>Export</TALLYREQUEST>
    <TYPE>Collection</TYPE>
    <ID>ID_Companies</ID>
  </HEADER>
  <BODY>
    <DESC>
      <STATICVARIABLES>
        <SVEXPORTFORMAT>$$SysName:UTF8JSON</SVEXPORTFORMAT>
        <SVCURRENTCOMPANY>Acme Retail Pvt Ltd</SVCURRENTCOMPANY>
      </STATICVARIABLES>
      <TDL>
        <TDLMESSAGE>
          <COLLECTION NAME="ID_Companies" ISMODIFY="No">
            <TYPE>Company</TYPE>
            <FETCH>NAME, STARTINGFROM, BOOKSFROM, STATENAME, PINCODE</FETCH>
          </COLLECTION>
        </TDLMESSAGE>
      </TDL>
    </DESC>
  </BODY>
</ENVELOPE>
```

Four core object streams are supported, each its own named `TYPE`/`FETCH` collection definition
(reader.go's `collectionDefs`):

- `companies` — `TYPE=Company`, fields `name`, `starting_from`, `books_from`, `state_name`,
  `pincode`. No incremental cursor (masters are always exported in full — legacy TallyPrime has no
  notion of a "company modified since" filter).
- `ledgers` — `TYPE=Ledger`, fields `name`, `parent` (the ledger's Group), `opening_balance`,
  `closing_balance`, `is_bill_wise_on`. Primary key `name` (TallyPrime ledger names are unique
  within a company).
- `groups` — `TYPE=Group`, fields `name`, `parent`, `is_revenue`, `is_deemedpositive`,
  `affects_gross_profit`. Primary key `name`.
- `stock_items` — `TYPE=StockItem`, fields `name`, `parent` (Stock Group), `base_units`,
  `closing_balance`, `closing_value`. Primary key `name`.
- `vouchers` — `TYPE=Voucher`, fields `guid`, `voucher_number`, `voucher_type`, `date`, `party_name`,
  `amount`, `narration`. Primary key `guid`. This is the one stream with an incremental cursor:
  `date` (`cursor_field` behavior mirrors `native/postgres`'s convention — the stored cursor is the
  last-seen voucher `date`, and `from_date`/`to_date` config, when set, are sent as
  `STATICVARIABLES.SVFROMDATE`/`SVTODATE` to bound the export server-side). `InitialState` seeds an
  empty cursor (full export); the persisted cursor advances via `connsdk.MaxCursor` after each
  `Read` the same way `native/postgres`'s cursor tracking does.

Master streams (companies/ledgers/groups/stock_items) have no incremental filter: TallyPrime's
Collection mechanism has no "modified since" concept for masters, so every `Read` re-exports the
full collection, matching the read-only, snapshot-only nature the quarantine finding describes.

## Write actions & risks

None. This is a read-only source connector — `capabilities.write` is `false` and `Write` always
returns `ErrUnsupportedOperation`. TallyPrime's Import-mode envelopes (`TALLYREQUEST=Import`) can
mutate company data (post vouchers, alter masters) but this connector never sends one; only
`TALLYREQUEST=Export` requests are built anywhere in this package. The read-side risk is scoped to
a local machine: every request targets `gateway_url`, which is expected to be a loopback/LAN
address for a TallyPrime instance under the operator's own control, not a public internet
endpoint — there is no cross-tenant or public-network exposure risk analogous to a hosted SaaS API
key leak.

## Known limits

- No REST resource surface: TallyPrime speaks one RPC-style envelope endpoint, not distinct
  paths/methods per object, so `api_surface.json` declares zero endpoints (matching
  `native/postgres`'s/`native/amazon-sqs`'s identical zero-endpoint Tier-3 precedent) rather than
  listing `covered_by`/`excluded` entries against a REST surface that does not exist.
- No CDC/streaming support: TallyPrime's Gateway Server is polled per `Read` call; there is no
  subscription/webhook mechanism, so no `ReadCDC`/`CDCReader` implementation exists here (unlike
  `native/postgres`'s documented pglogrepl stub) and `capabilities.cdc` is `false`.
- XML fallback mode (`envelope_format: xml`) is implemented and unit-tested but not the preferred
  path: `envelope_format: json` (TallyPrime 7.0+ native JSON export) is the documented default and
  is what production configuration should use whenever the target TallyPrime version supports it.
- Master streams (companies/ledgers/groups/stock_items) are always full snapshots; only `vouchers`
  supports incremental reads via a `date` cursor and optional `from_date`/`to_date` config bounds.
- `gateway_url` must point at a locally-reachable TallyPrime instance with its Gateway Server
  enabled — this connector cannot discover or start TallyPrime itself, and `Check` will fail (or,
  in `mode: fixture`, be skipped) if no TallyPrime process is listening.
- A `mode=fixture` config value short-circuits all network access (Check succeeds, Catalog returns
  canned streams, Read emits canned rows) — this is a test/conformance-harness affordance only and
  must never be set in production config, matching the postgres/amazon-sqs goldens' identical
  fixture-mode convention.
