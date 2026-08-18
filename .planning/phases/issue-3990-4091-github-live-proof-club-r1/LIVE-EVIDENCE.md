# Live GitHub evidence — Issues #3990 and #4091

**Observed:** 2026-08-15

**Provider:** GitHub.com

**Target:** the immutable run-owned repository recorded by
`.planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json`

**Credential carrier:** encrypted `pm` vault; added from stdin and referenced only by its local name

This is a sanitized evidence projection. It intentionally contains no credential material, approval
token, provider response body, external-proof request target, or rendered rate-scope value. The raw
external-proof artifact remains in the disposable project and is bound here only by SHA-256.

## #3990 — whole-surface certification

Production call chain:

`cmd/pm/main.go → cli.Run → app.Open / certification command registration → certify.Runner → definition-owned GitHub connector → rate-limit admission/observation → shared coordinator → GitHub`

The real-binary run used the production in-process certification route, `--full`, `--full-parity`,
`--write`, and `--external-proof`. Its non-secret rate-scope projection was supplied from a shell
variable and is deliberately not reproduced here. The pinned Node per-operation proof sweep was not
substituted: it requires a GitHub App installation and Polymetrics-Cert organization boundary that
were not present, and its own execution-model guard forbids it from claiming current in-process
certification.

| Observation | Result |
| --- | --- |
| Start / completion | `2026-08-15T08:50:43.053854Z` / `2026-08-15T08:54:12.519871Z` |
| Certification result | `passed=true` |
| Stages | 1,370 total; 1,214 passed; 156 explicitly skipped; 0 failed |
| Shared admission events | 83 attempts, 83 waits, 83 resets; 0 not-sent |
| Methods admitted | `DELETE`, `GET`, `POST` |
| Wait accounting | 447 ms total; 142 ms maximum |
| Provider throttling | no HTTP 429 and no abuse-throttle response |
| External child | exit 0; 83 fingerprinted exchanges |
| External status tally | 71×200, 1×201, 1×204, 6×403, 4×404; the 403/404 results are declared permission/unavailable terminal cases, not abuse throttling |
| External artifact | SHA-256 `715a9a707587c9c29746a08079d8c59052058ec677226e9f6dc83b5b7218f148` |
| Sanitized report | SHA-256 `a2480c68a86bbdc9dc3c80efbd002deb988de6085bec8ceb50226af329995b57` |
| Proof binary | SHA-256 `7ce8541c62575dc639d0e49efe94dd57ebf373414a74315185d8019624845f71` |
| Redaction | `repository_salted_hmac_sha256_v1`; one credential fingerprint; leak ledger `null` |
| Flow references | `flow_plan`, `flow_preview`, `flow_run`, `flow_status` |

The run also exposed and then proved three real defects:

- empty live streams were incorrectly treated as round-trip failures;
- query-contract stages ran after a typed unavailable read produced no capture;
- destructive cleanup tried to consume a token before preview and omitted typed destructive
  confirmation;
- the external proof guard applied a 16-exchange retry bound to the entire 83-exchange run rather
  than to one method/target pair.

Each received a focused RED before the minimal production fix. The final current run passed and an
independent `pm github label list` found zero labels matching its run tag.

### GraphQL live boundary

The fixed `graphql query rate-limit` command returned HTTP 200 with the provider's actual `cost`,
`limit`, `remaining`, and `resetAt` fields. Exact rate values are intentionally excluded from this
committed projection. A planned/previewed/approved fixed `addStar` mutation then changed the lab
repository's independently read `stargazers_count` from 0 to 1. A separately planned and approved
destructive `removeStar` returned it from 1 to 0. Both mutation results named `POST /graphql` and
HTTP 200; a fresh rate query after each mutation again returned an actual provider cost.

## #4091 — durable per-connection authorization

Production call chain:

`cmd/pm/main.go → cli.Run → app.Open → definition registry / transport construction → etl transport plan+preview → etl run dispatch → issueLabelDestinationExecutor → durable authorization gate → declarative GitHub PUT → independent GitHub read-back → acknowledged checkpoint`

The original live attempt exposed that `App.ApplyIssueLabelTransport` already supported an empty
token after durable authorization, but the shipped CLI parser rejected `--approval-plan` unless a
fresh token marker was also present. The focused RED failed with the old validation message. The
minimal fix keeps plan ID and `--confirm destructive` mandatory while making token stdin optional
only for later forward runs; cleanup still requires its separate single-use token.

The fixed real binary has SHA-256
`ccd3e295000afb4ca78a7d0ddfc87ff94f95ae8c99ffa16ae1c5f1c8931da238`.

| Mode / case | Observable live result |
| --- | --- |
| `full_overwrite` with `transport_allow_set_replace=true` | first approved run loaded 1 record; exact label read-back; identical-scope tokenless run loaded 1 record; exact read-back again |
| `incremental_upsert` with `transport_allow_keyed=true` | first approved run loaded 1 record; exact label read-back; identical-scope tokenless run loaded 1 record; exact read-back again |
| Old token replay, both modes | typed `AuthorizationTokenReplayError`; CLI validation refusal; provider label unchanged; stream checkpoint unchanged; failed run checkpoint absent |
| Real authentication refusal | typed `connsdk.HTTPError` status 401; provider labels unchanged; stream checkpoint unchanged; failed run loaded 0 and had no checkpoint |
| Cleanup | separately approved issue-label inverse plus separately approved repository-label deletes; final issue labels `[]` and run-prefix repository labels `[]` |

The required exact safe commands and real outputs were posted on
https://github.com/polymetrics-ai/cli/issues/4091.

## Captain edge-case matrix

| Edge case | #3990 | #4091 | Classification and negative assertion |
| --- | --- | --- | --- |
| Cancellation mid-operation | Shared admission cancellation and request cancellation are deterministic tests; deliberately terminating a live accepted GitHub request is not safely classifiable. | Deterministic orchestrator cancellation tests cover pre-apply and post-ack boundaries. | Simulated: pre-apply has zero provider send/checkpoint; post-ack preserves only the acknowledged checkpoint. |
| Connection/process dies partway | Fresh child process proved the external run boundary. Coordinator restart/persistence is deterministic because killing a live provider exchange would make send outcome unknowable. | Fresh-open, final-state-save-failure, and stale-writer tests cover interruption. | Mixed live/simulated: no hidden success claim; failed finalization stays typed and durable state remains truthful. |
| Empty input | Empty live streams now skip with a documented reason instead of inventing round-trip success. | Closed workset must contain exactly one reopened issue. | Live plus simulated refusal; zero destination write/checkpoint for invalid workset. |
| Single input | Multiple single-resource reads/writes passed. | Both modes loaded and read back exactly one record. | Live. |
| Large input | Whole run executed 1,370 stages and 83 provider exchanges. | The closed MVP route intentionally enforces batch size 1; a large batch is not executable. | Live whole-run boundary; typed CLI validation for #4091 and zero provider access. |
| Duplicate / out-of-order delivery | Parking coordinator duplicate/cancellation tests keep reservations observable without oversend. | Stale writer and second-page races cover duplicate/out-of-order checkpoint writers for all modes. | Simulated: losing writer cannot advance the stream checkpoint. |
| Schema drift | Certification fixture and schema checks are deterministic; corrupting GitHub's live schema is unavailable. | Receipt/artifact corruption and changed authorization scope refuse before provider write. | Simulated: typed/corruption error; zero provider write and no checkpoint advance. |
| Permission refusal | Six live 403 terminal cases were recorded and did not become retry storms; their scopes are fingerprinted, not rendered. | No separately under-scoped real credential was available. | #3990 live; #4091 explicitly untestable live. Deterministic consent/scope/revocation tests assert zero send. |
| Authentication refusal | Live credential passed; invalid authentication is covered by typed requester tests. | Real GitHub returned 401; labels and checkpoint were unchanged. | #4091 live typed `HTTPError`; zero records loaded and no checkpoint. |
| Concurrent same target | Multi-process tiny-budget test proves same-scope workers stay within the shared budget. | Destructive concurrent live PUTs would not reveal which provider write won. | #3990 simulated multi-process; #4091 simulated stale-CAS across all modes, loser checkpoint unchanged. |
| Resume after interruption | Coordinator restart resumes only after the stored reset; fresh child proof exits cleanly. | Fresh-open and acknowledged-checkpoint resume tests cover the durable boundary. | Simulated where forced death would be ambiguous; no duplicate unacknowledged checkpoint. |
| Replay acknowledged item | Strict mutation requester and proof-artifact exchange guards reject unsafe transport replay. | Both token replays were refused live with unchanged provider/checkpoint state. | #4091 live; typed error plus zero additional side effect. |

## Final cleanup ledger

- issue 5 labels: empty;
- repository labels with `pm-live-proof-4091-` prefix: empty;
- certification label tag: absent;
- repository star count: restored to 0;
- temporary remote fixture branch: deleted and absent on `ls-remote`;
- certification leak ledger: `null`;
- credential values or approval tokens observed in evidence: none.
