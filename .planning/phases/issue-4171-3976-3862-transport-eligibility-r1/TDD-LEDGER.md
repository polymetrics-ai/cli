# TDD ledger — transport source eligibility club

Status values begin `planned`; exact commands and retained traces are filled during execute-phase.

| ID | Guarantee | Red assertion | Green proof | Status |
| --- | --- | --- | --- | --- |
| R1 | GitHub source eligibility is an explicit executable-stream allowlist | Production-composed `commits` cannot pass the declaration, and the red test cannot compile until the typed admission contract exists. | GitHub declaration names every executable bundle stream and no wildcard; production composition accepts `commits`. | red |
| R2 | The allowlist still refuses absence before effects | `SourceStreamIneligibleError` is undefined in both registry and `app.Open` tests. | `errors.As` matches `SourceStreamIneligibleError`; source requests, stages, target sends/rows, and checkpoint updates all remain zero. | red |
| R3 | Declarative source pages rather than collecting one capped issue batch | The source is hard-coded to `issues`, ten pages, and 1,000 buffered records. | A multi-page `commits` fixture emits bounded pages, preserves all records, and binds each candidate to stream/ordinal/content. | planned |
| R4 | `max_pages` is honest | Omitted and unlimited settings cannot be distinguished by the transport adapter. | Omitted reads exactly one provider page; positive values cap; `0`/`all`/`unlimited` exhaust declared pagination. | planned |
| R5 | PostgreSQL polling is definition-selected in production | `app.Open` loads PostgreSQL with polling status `planned`, so its composed source cannot expose the exact shared executor reference. | Loaded PostgreSQL definition, `app.Open` factory composition, and native adapter resolve exact polling references/evidence and invoke the shared executor. | red |
| R6 | Resume uses strict `(cursor, complete primary-key tuple)` ordering | Existing scalar reader cannot represent duplicate watermark values or composite keys safely. | Live/native tests deliver duplicate-cursor, composite-key rows once, resume strictly after the tuple, and reject null/lossy/unsupported cursor/key shapes. | planned |
| R7 | Shared polling owns acknowledgement/checkpoint sequencing | A hand-built shared executor is the only proof and production PostgreSQL has no call chain. | Production route persists a candidate only after warehouse/target durable acknowledgement; failure/restart replays unacknowledged work and preserves acknowledged resume. | planned |
| R8 | Native page traversal is strict | Out-of-order or non-advancing native pages can reach the emitter in a faulty adapter. | Typed traversal refusal occurs before stage/send/checkpoint and leaves target/checkpoint unchanged. | planned |
| R9 | Schema/auth/generation drift fails closed | Production adapter has no dynamic fingerprint/key binding to validate. | Auth refusal performs no fetch; schema/key/source-generation drift returns typed rebootstrap/admission errors before downstream effects. | planned |
| R10 | Cancellation and concurrent-run fencing are safe | No composed-route proof exists. | Cancellation terminates bounded work without advancing the interrupted candidate; same-target concurrent runs obey lease/CAS fencing and do not double-commit. | planned |
| R11 | Empty, single, and large cardinalities are honest | Current GitHub adapter's cap hides large inputs and polling has no production route. | Empty sends zero/does not advance existing state; single sends one; large streams deliver the independently counted total in bounded pages. | planned |
| R12 | Acknowledged replay and duplicate/out-of-order delivery remain idempotent | No cross-family production proof. | Replayed acknowledged page does not add target rows; duplicate/out-of-order inputs follow declared keyed/ordered policies or fail typed with zero effects. | planned |
| R13 | Closed-spine conformance remains exact | API/database combinations can be hand-built without proving shipped composition. | `app.Open` loads exact references and accepted immutable evidence for API→database and database→database; mismatch/absence remains pre-I/O refusal. | planned |

## Initial red commands

```sh
go test -count=1 ./internal/synctransport -run 'Test.*SourceStream.*Ineligible'
go test -count=1 ./internal/app -run 'Test.*Transport.*(GitHub|Commits|Ineligible|Composition)'
go test -count=1 ./internal/connectors/native/postgres -run 'Test.*Polling.*(Definition|Transport|Resume|Refus)'
```

Initial failures and the corresponding green reruns are retained in `traces/`. A test that manually
registers the subject executor is supporting unit coverage only and cannot close R1, R5, R7, or
R13.
