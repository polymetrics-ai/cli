# TDD ledger — Issue 3886: dynamic schema discovery foundation

Every behavior-adding task starts red. `planned` becomes `RED` when exact
failing evidence is retained, then `GREEN` only after the same focused command
passes.

| ID | Guarantee | Red assertion | Status |
| --- | --- | --- | --- |
| D1 | Static/discovered equivalence | A static raw bundle schema and constructed discovered schema cannot reach one catalog stream constructor. | GREEN |
| D2 | Sync contract | `x-primary-key` / `x-cursor-field` from each schema produce the same fields and derived catalog metadata. | GREEN |
| D3 | Bounded work | More than ten descriptions can run concurrently. | GREEN |
| D4 | Cancellation | Cancelled context leaves no blocked worker/result send. | GREEN |
| D5 | Rate limits | A typed 429 is not retried with bounded exponential jitter. | GREEN |
| D6 | Progress | A 100+ object discovery emits monotonic 100-item heartbeat status without provider/error text. | GREEN |
| D7 | Honest partial result | Failed global enumeration returns no fallback; failed description reports a falsely complete catalog. | GREEN |
| D8 | Account cache | Two connections for one opaque coordination identity fail to reuse a fresh cache; different opaque identities share one. | GREEN |
| D9 | Explicit refresh | Explicit refresh fails to bypass a fresh durable or in-process schema cache. | GREEN |
| D10 | Secret safety | Cache/status/error serialization can contain a runtime secret, raw credential name, or raw error body. | GREEN |
| H1 | Unknown HubSpot object | A fixture custom type/property unknown to source cannot be listed and schema-generated. | GREEN |
| H2 | HubSpot properties | `bool`, `number`, `date`, `datetime`, enumeration/options, reference type and unique metadata do not map truthfully. | GREEN |
| H3 | HubSpot fallback | Schema-list failure loses declared standard baseline rather than marking fallback/partial. | GREEN |
| H4 | HubSpot rate limit | A `429` discovery response is not retried by the shared driver. | GREEN |
| H5 | HubSpot read | A cataloged fixture custom type cannot execute its fixed object collection path/pagination, or emits fields not in discovery schema. | GREEN |
| A1 | Durable account cache | Stored catalogs are inline, key on a raw credential, or duplicate the same opaque account across connections. | GREEN |
| A2 | Stale display | `pm catalog show` human and JSON output omit the stale/refresh-required state. | GREEN |
| A4 | Storage boundary | Catalog callers reach `state.Catalogs` directly, so account files cannot serve future destinations. | GREEN |
| A5 | Disk safety | A persisted catalog file contains runtime config, secret values, raw credential names, binding IDs, or a cache key. | GREEN |
| A6 | Crash ordering | State can point to a catalog file before the file and created directory chain have synced. | GREEN |
| A3 | Refresh behavior | `pm catalog refresh` fails to bypass a fresh cache. | GREEN |

## Planned red commands

```sh
go test ./internal/connectors ./internal/connectors/discovery ./internal/connectors/engine -run 'Test(Catalog|Discovery|Schema)' -count=1
go test ./internal/connectors/native/hubspot ./internal/connectors/native/nativeset ./internal/connectors/bundleregistry -run 'TestHubSpot|TestFactories|TestNew' -count=1
go test ./internal/app ./internal/cli -run 'Test(Catalog|RunCatalog)' -count=1
```

No live provider check belongs in this ledger; all provider fixtures are local
and contain only fabricated object/property identifiers.
