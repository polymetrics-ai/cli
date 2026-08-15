# TDD ledger — #4093

## Planned RED → GREEN checkpoints

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Loader/projection | `go test -timeout 20m -run 'TestBundleLoadSyncTransport|TestDestinationTransportDescriptorRefusesChangeCapture' ./internal/connectors/engine ./internal/connectors` failed: the loaded `Definition().SyncTransport` was nil and every unknown/missing/wrong-version declaration was ignored. | The same command passes: versioned loader plus clone-safe `Bundle`/`Definition` projection makes a valid declaration observable while `additionalProperties: false`, strict decoding, version checks, and descriptor validation refuse the malformed cases. | green |
| Atomic definition composition | `go test -timeout 20m -run 'TestRegisterDeclaredTransports|TestDefinitionConformanceVerifier' ./internal/synctransport` did not compile because no definition composer or factory authority existed. | The same command passes: valid declarations build and register two source/destination pairs; unknown executor, invalid destination role, and altered evidence all observe zero builders, zero registrations, and zero source reads. | green |
| Production registrations | `TestOpenRegistersDefinitionOwnedProductionTransports` was absent while App composed only a GitHub wrapper. | The app-open preflight test observes the PostgreSQL `postgres_bounded_snapshot` source and GitHub `issue_label_destination` destination registered from their production `sync_transport.json` declarations. | green |
| Destination role rule | `go test -count=1 -timeout 20m ./internal/connectors -run TestDestinationTransportDescriptorAllowsChangeCaptureWithClosedApplyStrategy` fails because a destination that declares the closed `change_apply` strategy is refused before dispatch. | The connector test and `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination` accept a legitimate `change_capture` destination, reach the exact `change_apply` plan and one apply; a malformed `change_capture`/`append` declaration still refuses before registration or I/O. | green |
| PostgreSQL live proof | Existing live native source test exercises current Go-owned descriptor. | Pending the mandated Docker/Colima run against the now definition-owned descriptor and production factory. | planned |
| Connector boundary repair | PR CI rejected the initial direct `internal/app` import of `native/postgres` as a connector-boundary violation. | `DefinitionFactoriesFromRegistry` gathers connector-provided factories without an App connector import; the production preflight test still resolves PostgreSQL and `go run ./cmd/connectorgen boundary . --json` passes. | green |
| Embedded definition repair | Full CI RED observed GitHub and PostgreSQL definitions with no transport declaration; the binary route refused the issue-label approval and native PostgreSQL tests found no source descriptor. | `defs.FS` embeds `*/sync_transport.json`; `TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle` and the PostgreSQL definition/registration tests pass, making the roles observable in the production binary. | green |
| Inspection projection repair | Full CI RED found `connectors inspect github --json` still asserted both transport roles were unsupported after the GitHub bundle declared them. | The CLI projection test observes `source.status=declared` and `destination.status=declared` from the production definition; runtime help, CLI manual, generated connector docs, and website guide explain the visible status. | green |
| Website data repair | Website CI RED detected the generated agent-guide data was stale after the transport declaration guidance changed. | `pnpm run gen:website-data` updates only `website/lib/docs.generated.ts`; the generated page includes the GitHub declared-role explanation. | green |
| CI regression repair | `go test -count=1 -timeout 20m ./internal/app -run 'TestGithubPullRequestsETLSupportsLegacyExecutableModes|TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume'` failed: the noncanonical GitHub mode ordering bypassed the exact issue-label route guard and a generic destination fixture declared forbidden `change_capture`. | The focused app command and `go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` pass: GitHub’s declaration is canonical, so its one-sided legacy warehouse connection remains legacy; test-only destinations advertise only modes the shared descriptor permits; the generated CLI transcript records the declared roles. | green |

Every refusal test asserts the relevant side effect count is zero. No test relies
only on `err != nil` or a lack of panic.

## Red:

The loader/projection test failed against the base because the transport field
was absent; the composition test did not compile because the factory/composer
API did not exist; and the destination rule initially refused every
`change_capture` destination, including a declaration with the closed
`change_apply` strategy.

The post-push CI run failed after the production declaration activated the
existing app route guard. The focused app reproduction observed the GitHub
descriptor but rejected it as an issue-label route solely because its declared
mode order differed from `synccontract.AllModes()`; independently, the generic
destination fixture exposed that the role rule refused a legitimate closed
`change_capture` route.

## Green:

The same focused test commands now pass with a versioned bundle loader,
definition projection, prevalidated atomic registration, factory-held evidence
authority, production GitHub/PostgreSQL declarations, and a destination-role
refusal before any executor state changes.

The CI-regression focused app command now observes GitHub-to-warehouse using
the legacy executable path. `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination`
observes one source read, warehouse stage, destination plan, and destination
apply with `change_apply`; the malformed declaration composition test observes
zero builds, registrations, reads, plans, and applies. The regenerated CLI
golden transcript also records the observable declared GitHub source and
destination roles.
