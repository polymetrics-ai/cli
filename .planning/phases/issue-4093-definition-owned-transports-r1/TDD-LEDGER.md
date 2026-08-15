# TDD ledger — #4093

## Planned RED → GREEN checkpoints

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Loader/projection | Tests expect a `sync_transport.json` descriptor, strict failures, and independent definition projections; current engine has no field/loader. | Versioned loader plus clone-safe `Bundle`/`Definition` projection. | planned |
| Atomic definition composition | Tests count build/registration side effects and expect zero for malformed/unknown declarations; no composition API exists. | Valid declarations construct registered adapters only after full prevalidation. | planned |
| Production registrations | App-open tests expect declared GitHub and PostgreSQL roles to preflight through real adapters; current App composes only a GitHub wrapper. | Reference-indexed GitHub/PostgreSQL factories and definition-owned JSON declarations. | planned |
| Destination role rule | Destination `change_capture` test expects structural refusal; existing descriptor accepts it. | Destination validation rejects it before registration or I/O. | planned |
| PostgreSQL live proof | Existing live native source test exercises current Go-owned descriptor. | Same real rows/pages/checkpoint through the definition-owned descriptor and production factory. | planned |

Every refusal test asserts the relevant side effect count is zero. No test relies
only on `err != nil` or a lack of panic.
