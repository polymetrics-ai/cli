# Verification — Zoom runnable command surface

Issue: #4265
Phase: `zoom-full-definition-mapping-r1`

## Result

CURRENT BRANCH HAS A FOUNDATION-DEPENDENT RED; FINAL FULL AND LIVE GATES PENDING. The historical
full-repository and observed-read evidence below predate the current 206-action surface and does not
satisfy the required credentialed-certification gate. No Zoom capability cell is certified: the
matrix requires an exact fixture plus live proof, and the final built-binary/App path has not yet
been exercised. The optional-query Foundation rehearsal at `c3f83cbf6eabbae00219566fb02719ca2d6c480d`
is now incorporated through `main`. The retained-artifact contract now verifies 34 re-pinned Zoom
OpenAPI documents; the exact remaining source boundary is the explicit unavailable Accounts capture.

## Focused evidence

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestMissingFoundationGapRowsAreSourceLockedAndRollUp|TestZoomTransportDeclaresTheExecutableSourceOnlyUntilDeleteSemanticsExists|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable)$'` | PASS (2026-08-23) — source-only ETL transport, all eight implemented direct CLI DELETE commands, the zero-destination readiness total, and the source-traced delete-semantics gap are internally coherent. |
| `go test -count=1 -timeout 20m ./internal/connectors/conformance -run '^TestConformance/zoom$'` | PASS (2026-08-24, 4.615s) — the Zoom bundle's fixture conformance remains green after source retention and canonical-flag reconciliation. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestProviderInventoryLedgerIsComplete|TestMissingFoundationGapRowsAreSourceLockedAndRollUp|TestDeclarationDispositionAccountsForThePinnedSourceAndLedger|TestEveryTypedZoomActionHasReverseETLCommandAndCandidate|TestZoomTransportDeclaresTheExecutableSourceAndTypedDestination|TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestRunnableOperationContractsHaveCommands)$'` | PASS (2026-08-22) — all current source-locked inventory, declaration-disposition, missing-foundation, command/action, transport, candidate, and seven-surface mapping assertions pass. The intentionally Foundation-dependent lifecycle test is excluded. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestProviderInventoryLedgerIsComplete|TestDeclarationDispositionAccountsForThePinnedSourceAndLedger|TestEveryTypedZoomActionHasReverseETLCommandAndCandidate|TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites|TestRunnableOperationContractsHaveCommands|TestZoomTransportDeclaresTheExecutableSourceAndTypedDestination)$'` | Historical PASS before the current Meeting DELETE optional-query regression; it must be rerun after Foundation reaches `main`. It is not current lifecycle evidence. |
| `go test -timeout 20m ./internal/connectors/defs/zoom -count=1` | Historical PASS — 1,748 source contracts, 311 delete contracts, 712 runnable commands, and the exhaustive no-credential preflight sweep. It must be rerun for this surface. |
| `go test -timeout 20m ./internal/connectors/defs/zoom -run '^TestEveryTypedZoomActionHasReverseETLCommandAndCandidate$' -count=1` | Historical PASS — 204 implemented reverse-ETL commands, typed write actions, and generated mutation candidates are an exact one-to-one set. The #4304 continuation adds the closed destination declaration. |
| `go test -timeout 20m ./internal/connectors/engine -run '^TestEveryShippedWriteActionHasExpectedBatchability$' -count=1` | PASS — Zoom actions use the repository default batchability policy while retaining reverse-ETL approval. |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1` | PASS — regenerated root-help transcripts for the declared Zoom surface. |
| `go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1` | PASS — regenerated transcripts are asserted without update mode. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json` | PASS (2026-08-22) — no Zoom definition findings. |
| `go run ./cmd/connectorgen certification-matrix . --connector zoom --check` | PASS (2026-08-24) — connector shard is current; one connector, zero capability-complete and zero certified claims. |
| `go run ./cmd/connectorgen certification-candidates . --connector zoom && go run ./cmd/connectorgen certification-candidates . --connector zoom --check` | PASS (2026-08-24) — the canonical-flag repair intentionally made the command-derived candidate stale; it was regenerated and rechecked without claiming a destination capability. |
| `go run ./cmd/connectorgen certification-sweep . --connector zoom && go run ./cmd/connectorgen certification-sweep . --connector zoom --check` | PASS (2026-08-24) — regenerated and current: 717 rows and 714 CLI commands. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json` | Historical BLOCKED (2026-08-23) — v2 lock parsing rejected `retrieval`; resolved by #4332's rendered-reference contract. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestPinnedSourceCrosswalkAccountsForEveryIdentity|TestMissingFoundationGapRowsAreSourceLockedAndRollUp)$'` | PASS (2026-08-24) for the provisional v3 rendered-reference conversion: 35 documents, 12,127,228 bytes, and 1,937 operation identities retained. The probe was intentionally not retained after the next source-import failure. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json` | Expected intermediate BLOCKED (2026-08-24) — v3 lock parsing succeeded; the next refusal was a missing canonical `sources/zoom-operation-descriptor.json`. |
| `go run ./cmd/connectorgen source-import zoom --out internal/connectors/defs/zoom/sources/zoom-operation-descriptor.json` | Terminal BLOCKED (2026-08-24) — the first immutable capture artifact, `accounts`, returned HTTP 404; no verified cache entry exists. The rendered-reference contract deliberately does not treat cited provenance as fetched bytes. The original v2 lock was restored byte-for-byte (`SHA-256 2e102acffd89467374405829abd994714f994f237c4a38c4ad0d9a553c42c3f7`) pending the separately decided stable, attested capture mirror. |
| `go run ./cmd/connectorgen source-retain zoom --retrieved-at <recorded UTC> --license undetermined --terms undetermined` | PASS (2026-08-24) — 34 exact current first-party OpenAPI artifacts were retained (11,719,368 bytes) with a connector-owned provenance manifest. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestPinnedSourceCrosswalkAccountsForEveryIdentity|TestMissingFoundationGapRowsAreSourceLockedAndRollUp)$'` | PASS (2026-08-24) — v3 source ownership, all 34 authorized re-pins, retained-gap provenance, 1,871 current operations, 66 historic unavailable Accounts identities, and the unchanged 1,937-operation crosswalk are coherent. |
| `go run ./cmd/connectorgen source-import zoom --check` | Expected BLOCKED (2026-08-24) — source import refuses exactly the declared Accounts unavailability before a network/cache fallback or descriptor projection. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json` | Expected BLOCKED (2026-08-24) — one `source_projection` finding repeats the exact Accounts unavailable reason. |
| `go run ./cmd/connectorgen surface-sync --check` | Expected BLOCKED (2026-08-24) — canonical source descriptor is absent because the Accounts unavailable declaration correctly prevents descriptor generation. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom` | PASS (2026-08-24, 654.179s) — complete package suite, including every implemented command's no-credential preflight sweep. The prior duplicate path-parameter flag drift is repaired without changing command availability. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable$'` | PASS (2026-08-24, 1.386s) — the optional-query Foundation now executes the exact source-local Meeting lifecycle fixture shapes; no live provider mutation or certification claim. |
| `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestZoomMeetingDeleteOptionalQueryRehearsal$'` in isolated detached SHA `c3f83cbf6eabbae00219566fb02719ca2d6c480d` | PASS — exact Zoom Meeting DELETE query declaration sends the present optional field and omits both absent optional record fields on a fixture-approved loopback request. SHA-bound behavior only; no credentials, Zoom call, certification, or connector-branch ancestry. See [`FOUNDATION-REHEARSAL.md`](FOUNDATION-REHEARSAL.md). |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestMissingFoundationGapRowsAreSourceLockedAndRollUp$'` | Historical PASS before the delete-only reconciliation — 12 deduplicated gap IDs. The current focused pass above proves 11 catalog gaps, including the eight-row delete-semantics fan-out, with the same 1,329 source-locked rows and 1,299 affected provider operations. |
| `go run ./cmd/connectorgen surface-sync --check` | PASS (2026-08-22) — zero generated field drift. |
| `go build ./cmd/pm` | PASS (2026-08-24) — the installed binary builds with the reconciled source and command declarations. |
| `make docs-check` | PASS (2026-08-24) — regenerated CLI/Zoom connector manuals and catalog validate. |
| `go run ./cmd/agentcontractgen check` | PASS (2026-08-24) — canonical contract and registered projections are current. |
| `make connector-boundary` | PASS (2026-08-24) — whole-tree outcome `clean`; 317 files and 553 connector bundles checked; only the repository's pre-existing, expiring non-Zoom exceptions were reported. |

## Publication audit (2026-08-22)

- The temporary #4304 stacked-base history has landed through `main`; PR #4285 is a draft to
  `main`. The exact optional-query Foundation rehearsal SHA
  `c3f83cbf6eabbae00219566fb02719ca2d6c480d` is now incorporated through that mainline history.
- The pending cohort is limited to `internal/connectors/defs/zoom/**` and the matching
  `.planning/phases/zoom-full-definition-mapping-r1/**` lifecycle evidence. The disposable
  `.zoom-foundation-rehearsal-c3f83` worktree has been removed; no runtime artifact, recovery
  sentinel, or unrelated path is tracked or untracked in this worktree.
- A filename-only credential-pattern scan over branch-ahead, modified, and untracked cohort files
  returned no matches, and `git diff --check` is clean. The scan records no secret values.

## Full gate

Historical `make verify` — PASS after disk recovery. It completed `gofmt`, `go mod tidy`, `go vet ./...`,
`go test -timeout 20m ./...`, `go build ./cmd/pm`, connector docs validation, smoke, lint,
agent-contract check, definition validation, surface sync, certification artifact checks, connector
boundary, connector canon, pinned build dependencies, Homebrew notification, and release-target
parity checks. The all-package test reported `internal/cli` PASS in 614.326s and the Zoom bundle
PASS.

## Command/certification boundary

The branch exposes 714 commands: 505 direct reads, three preserved ETL streams, and 206 guarded
typed-write commands, including 185 DELETE actions. All are implemented pending certification.
The eight actions with `users.id -> user_id` overlap remain directly CLI-reachable but are not sync
destinations: ordinary replay would be a destructive delete and is refused at
`internal/app/issue_label_warehouse_transport.go:944`. The connector-local matrix records no
current live-tested `operation:rest_read` cell and zero certified cells. The retained 2026-08-19
proof is historical because it has no current certification subject.

## Live certification evidence

| Candidate/stage | Result | Honest conclusion |
| --- | --- | --- |
| `api users user` bounded authenticated direct read | Historical PASS — two HTTP 200 exchanges, retained as fingerprint-only `observed_operations` evidence | Not current proof and **not certified**: its record lacks the current certification subject; a fresh matching proof also needs `operation-specific-fixture-evidence-projection`. |
| `users` full-refresh append plus query read-back | PASS within the bounded full run | Uncertified: the hard-wired fixture-conformance skip and aggregate-report policy prevent publication of an accepted per-capability record. |
| `meetings` full-refresh append plus query read-back | PASS within the bounded full run | Uncertified: the hard-wired fixture-conformance skip and aggregate-report policy prevent publication of an accepted per-capability record. |
| `webinars` full-refresh append | NOT CERTIFIED — Zoom returned HTTP 400 for `GET /v2/users/me/webinars` | Recorded as a provider refusal without retaining its response body. |
| three preserved streams catalog acceptance | PASS | The SHA-pinned provider schemas give each stream a real creation timestamp, projected as `created_at`; no synthetic watermark is used. |
| 206 typed mutations / 185 deletes | NOT CERTIFIED | All remain directly CLI-reachable and explicitly unassessed. The eight `users.id → user_id` overlaps are provider DELETEs; none is a sync destination until a tombstone-aware shared delete transport exists. No provider mutation was attempted. |
| Final required proof | PENDING | After #4331 lands and the source lock validates, run the built CLI and persisted App path with only the registered secret-store reference at execution time: authenticated read; lane-owned Meeting create/read-back/update/delete cleanup; ETL; every supported destination plan/apply/acknowledgement and independent read-back; and a documented supported binary round-trip. |

The Webinar response was HTTP 400 with Zoom error code 200: Webinar plan is missing; subscribe to
the Webinar plan and enable webinars for the user before the action can run. The account identifier
from that response is redacted. This historical observation is not certification under the current
captain requirement and cannot replace the final registered-secret-store execution proof. No credential
value, token, raw provider response body, or account identifier appears in the repository, command
arguments, evidence, or this record.

The rerun contained 52 passing and 32 non-passing stages. The exact blockers to accepting its two
passing source capabilities are `definition-fixture-conformance-certification-stage`
(`internal/connectors/certify/stages_source.go:811-814`),
`capability-scoped-live-evidence-publication` (`:673-695`, `:431`), and
`schedule-roundtrip-source-only-skip` (`:690-694`). They are connector-neutral foundation work;
no engine/auth/certification-status code is modified here.
