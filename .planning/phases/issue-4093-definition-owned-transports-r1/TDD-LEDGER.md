# TDD ledger — #4093

## Planned RED → GREEN checkpoints

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Loader/projection | `go test -timeout 20m -run 'TestBundleLoadSyncTransport|TestDestinationTransportDescriptorRefusesChangeCapture' ./internal/connectors/engine ./internal/connectors` failed: the loaded `Definition().SyncTransport` was nil and every unknown/missing/wrong-version declaration was ignored. | Versioned loader plus clone-safe `Bundle`/`Definition` projection. | red |
| Atomic definition composition | Tests count build/registration side effects and expect zero for malformed/unknown declarations; no composition API exists. | Valid declarations construct registered adapters only after full prevalidation. | planned |
| Production registrations | App-open tests expect declared GitHub and PostgreSQL roles to preflight through real adapters; current App composes only a GitHub wrapper. | Reference-indexed GitHub/PostgreSQL factories and definition-owned JSON declarations. | planned |
| Destination role rule | The same RED run observed `DestinationTransportDescriptor.Validate() = nil` for `change_capture`; no execution was attempted. | Destination validation rejects it before registration or I/O. | red |
| PostgreSQL live proof | Existing live native source test exercises current Go-owned descriptor. | Same real rows/pages/checkpoint through the definition-owned descriptor and production factory. | planned |

Every refusal test asserts the relevant side effect count is zero. No test relies
only on `err != nil` or a lack of panic.
