# TDD ledger — Issue #3992 schedule authorization

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Safe durable binding | a schedule has no durable authorization field or accepts token-like material | create requires an opaque `auth_<id>` reference; manifest/state tests reject non-reference material and search persisted files for fixture secret/token |
| R2 | Rendered payload carries no authority material | a backend payload omits the reference or contains secret/token material | all backend renderers invoke `schedule fire` with a quote-safe opaque reference and test searches the full payload |
| R3 | Unattended authorized firing | an installed payload still requires an approval token | running the crontab payload writes and reads back through the fixture connector with no token supplied |
| R4 | Scope safety | drift can reach target validation, write, or read | changed scope produces `AuthorizationScopeChangedError` with zero target events |
| R5 | Revocation / expiry | revoked or expired binding can dispatch or loses typed reason | each returns its exact app error and target counters remain unchanged |
| R6 | No replay | overlap, interruption, ambiguous failure, or rate limit can execute again | persistent fire state is halted/parked and a subsequent firing is rejected before a target write |
| R7 | Terminal evidence | a completed firing lacks terminal state or receipt link | status has terminal flow result and opaque receipt IDs after write/read-back; backend cleanup restores fixture bytes |
| R8 | Existing certification remains authorization-aware | the sample certification creates a schedule without the new required reference | the legacy non-action round-trip passes only an opaque reference; its real sample route and the dedicated authorized-fire fixture both pass |

## Red command

```sh
go test -timeout 20m ./internal/schedule ./internal/cli -run 'TestSchedule(Manifest|Fire|CLI|Render)' -count=1
```

The first red output is retained at `traces/red-run.txt` before production changes.

## Green command

```sh
go test -timeout 20m ./internal/schedule ./internal/cli -run 'TestSchedule|TestManifest|TestInstalledScheduleFire' -count=1
# PASS — schedule and CLI focused suites
```

The final package and repository gates are recorded in `VERIFICATION.md`.

`traces/certification-red.txt` records the compatibility red stage and its
green proof (`TestCertifyCLISingleConnectorPassExitsZero`).
