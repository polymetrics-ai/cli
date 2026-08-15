# TDD ledger — transport source eligibility club

## Scope correction — 2026-08-16

PR 4175 owns #3976. This branch's attempted PostgreSQL adapter was removed after the real CLI
inspection guard showed that polling was advertised before a shipped production preflight could
bind its dynamic contract. The retained guard is the red/green evidence; it is not loosened.

Status values begin `planned`; exact commands and retained traces are filled during execute-phase.

| ID | Guarantee | Red assertion | Green proof | Status |
| --- | --- | --- | --- | --- |
| R1 | GitHub source eligibility is an explicit executable-stream allowlist | Production-composed `commits` cannot pass the declaration, and the red test cannot compile until the typed admission contract exists. | GitHub declaration names every executable bundle stream and no wildcard; `app.Open` composition accepts `commits`. | green — `TestOpenRegistersDefinitionOwnedProductionTransports` |
| R2 | The allowlist still refuses absence before effects | `SourceStreamIneligibleError` is undefined in both registry and `app.Open` tests. | `errors.As` matches `SourceStreamIneligibleError`; source requests, stages, target sends/rows, and checkpoint updates all remain zero. Case-equivalent `ISSUES` is also refused. | green — `TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess` |
| R3 | Declarative source pages rather than collecting one capped issue batch | The source is hard-coded to `issues`, ten pages, and 1,000 buffered records. | A multi-page `commits` fixture emits five bounded batches containing all 103 records and binds each candidate to stream/ordinal/content. | green — `TestOpenComposedGitHubCommitsSourceEmitsEveryUnlimitedPageInBoundedBatches` |
| R4 | `max_pages` is honest | Omitted and unlimited settings cannot be distinguished by the transport adapter. | Omitted reads exactly one provider page; positive values cap; `0`/`all`/`unlimited` exhaust declared pagination. | green — `TestOpenComposedGitHubCommitsHonorsTransportMaxPages` |
| R5 | PostgreSQL polling remains planned until a production preflight can bind it | `go test -v -count=1 -timeout 20m ./internal/cli -run '^TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt$'` failed because inspection reported `implemented`. | The unchanged CLI guard now observes `status=planned`, a non-empty reason, and no executable polling contract after the declaration was restored. | green — focused real-CLI guard |
| R13 | Closed-spine conformance remains exact | API/database combinations can be hand-built without proving shipped composition. | `app.Open` loads exact references and accepted immutable evidence for API→database and database→database; mismatch/absence remains pre-I/O refusal. | green — production-composition tests |
| G1 | Certification guards track the declaration-owned source reference exactly | At merge base `ef3c71caf`, both certification guards pass; at this branch's `73280ed81` head they fail because they still expect the replaced `issue_label_source` ID. | The assertions now require `declarative_stream_source` both for successful resolution and the deliberately unregistered typed refusal. Neither guard is removed or weakened. | green — `TestCertificationDeclaredTransportPair*` |

## Initial red commands

```sh
go test -count=1 ./internal/synctransport -run 'Test.*SourceStream.*Ineligible'
go test -count=1 ./internal/app -run 'Test.*Transport.*(GitHub|Commits|Ineligible|Composition)'
go test -v -count=1 -timeout 20m ./internal/cli -run '^TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt$'
```

Initial failures and the corresponding green reruns are retained in `traces/`. A test that manually
registers the subject executor is supporting unit coverage only and cannot close R1 or R13. R5 is
closed only by the real CLI inspection projection.

## Execution record

- **Red:** retained in `traces/red-stream-admission.txt` and
  `traces/red-postgres-polling-production.txt` before the production changes.
- **Green:** focused package, race, vet, build, generation, and repository drift gates are recorded
  in `VERIFICATION.md`; live GitHub certification is explicitly pending under the task's known
  external limits. PostgreSQL polling implementation is deliberately absent and owned by PR 4175.
