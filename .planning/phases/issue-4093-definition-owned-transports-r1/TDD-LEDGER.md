# TDD ledger — #4093

## Planned RED → GREEN checkpoints

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Loader/projection | `go test -timeout 20m -run 'TestBundleLoadSyncTransport|TestDestinationTransportDescriptorRefusesChangeCapture' ./internal/connectors/engine ./internal/connectors` failed: the loaded `Definition().SyncTransport` was nil and every unknown/missing/wrong-version declaration was ignored. | The same command passes: versioned loader plus clone-safe `Bundle`/`Definition` projection makes a valid declaration observable while `additionalProperties: false`, strict decoding, version checks, and descriptor validation refuse the malformed cases. | green |
| Atomic definition composition | `go test -timeout 20m -run 'TestRegisterDeclaredTransports|TestDefinitionConformanceVerifier' ./internal/synctransport` did not compile because no definition composer or factory authority existed. | The same command passes: valid declarations build and register two source/destination pairs; unknown executor, invalid destination role, and altered evidence all observe zero builders, zero registrations, and zero source reads. | green |
| Production registrations | `TestOpenRegistersDefinitionOwnedProductionTransports` was absent while App composed only a GitHub wrapper. | The app-open preflight test observes the PostgreSQL `postgres_bounded_snapshot` source and GitHub `issue_label_destination` destination registered from their production `sync_transport.json` declarations. | green |
| Destination role rule | The same RED run observed `DestinationTransportDescriptor.Validate() = nil` for `change_capture`; no execution was attempted. | Destination validation rejects it before registration or I/O. | green |
| PostgreSQL live proof | Existing live native source test exercises current Go-owned descriptor. | Pending the mandated Docker/Colima run against the now definition-owned descriptor and production factory. | planned |
| Connector boundary repair | PR CI rejected the initial direct `internal/app` import of `native/postgres` as a connector-boundary violation. | `DefinitionFactoriesFromRegistry` gathers connector-provided factories without an App connector import; the production preflight test still resolves PostgreSQL and `go run ./cmd/connectorgen boundary . --json` passes. | green |
| Embedded definition repair | Full CI RED observed GitHub and PostgreSQL definitions with no transport declaration; the binary route refused the issue-label approval and native PostgreSQL tests found no source descriptor. | `defs.FS` embeds `*/sync_transport.json`; `TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle` and the PostgreSQL definition/registration tests pass, making the roles observable in the production binary. | green |
| Inspection projection repair | Full CI RED found `connectors inspect github --json` still asserted both transport roles were unsupported after the GitHub bundle declared them. | The CLI projection test observes `source.status=declared` and `destination.status=declared` from the production definition; runtime help, CLI manual, generated connector docs, and website guide explain the visible status. | green |
| Website data repair | Website CI RED detected the generated agent-guide data was stale after the transport declaration guidance changed. | `pnpm run gen:website-data` updates only `website/lib/docs.generated.ts`; the generated page includes the GitHub declared-role explanation. | green |

Every refusal test asserts the relevant side effect count is zero. No test relies
only on `err != nil` or a lack of panic.

## Red:

The loader/projection test failed against the base because the transport field
was absent; the composition test did not compile because the factory/composer
API did not exist; and the destination rule returned success for
`change_capture`.

## Green:

The same focused test commands now pass with a versioned bundle loader,
definition projection, prevalidated atomic registration, factory-held evidence
authority, production GitHub/PostgreSQL declarations, and a destination-role
refusal before any executor state changes.
