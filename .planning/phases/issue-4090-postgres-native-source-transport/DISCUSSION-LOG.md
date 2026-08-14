# Discussion log — Issue #4090

Manual inline `discuss-phase` record, 2026-08-14.

| Question | Decision | Evidence |
| --- | --- | --- |
| What is selected? | The PostgreSQL connector's `Definition().SyncTransport` declares one exact `native_database` source executor. | #4090 acceptance; `connectors.SyncTransportDescriptorOf`; `synctransport.Registry.Preflight` |
| What is reused? | Typed catalog/type projection, bounded resource policy, deterministic read-plan constraints, and PostgreSQL connection/TLS code from #3976/#3974. | `native/postgres/{typed_catalog,catalog_types,connection}.go`; `database/read_plan.go` |
| Which modes? | Only `full_overwrite` and `full_append`; no polling/keyset/incremental or CDC promotion. | #4090 scope; #3858 and #3977 exclusions |
| How do failures behave? | Missing descriptor, incompatible database executor family, and unregistered exact executor reject at registry preflight before any pool/query I/O. | `synctransport.Registry.Preflight` ordering; #4090 acceptance |
| What proves reality? | Unit tests prove each pre-I/O refusal separately; a PostgreSQL 16.10 dbtest fixture proves bounded rows, identity/schema, and checkpoint output. | #4090 launch brief; existing `dynamic_catalog_integration_test.go` harness |

No product decision is unresolved. The implementation must stop rather than
add a shared transport contract, raw SQL surface, target support, generic
poller, or certification-schema change.
