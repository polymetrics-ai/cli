# Verification — Zoom runnable command surface

Issue: #4265
Phase: `zoom-full-definition-mapping-r1`

## Result

CURRENT BRANCH HAS A FOUNDATION-DEPENDENT RED; FINAL FULL AND LIVE GATES PENDING. The historical
full-repository and observed-read evidence below predate the current 206-action surface and does not
satisfy the required credentialed-certification gate. No Zoom capability cell is certified: the
matrix requires an exact fixture plus live proof, and the final built-binary/App path has not yet
been exercised. The optional-query Foundation rehearsal at `c3f83cbf6eabbae00219566fb02719ca2d6c480d`
is now incorporated through `main`; current validation is instead blocked on #4331's rendered
reference citation contract for the preserved 35 Zoom captures.

## Focused evidence

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestMissingFoundationGapRowsAreSourceLockedAndRollUp|TestZoomTransportDeclaresTheExecutableSourceOnlyUntilDeleteSemanticsExists|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable)$'` | PASS (2026-08-23) — source-only ETL transport, all eight implemented direct CLI DELETE commands, the zero-destination readiness total, and the source-traced delete-semantics gap are internally coherent. |
| `go test -count=1 -timeout 20m ./internal/connectors/conformance -run '^TestConformance/zoom$'` | PASS (2026-08-23) — the Zoom bundle's fixture conformance remains green after removing the invalid destination declaration. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestProviderInventoryLedgerIsComplete|TestMissingFoundationGapRowsAreSourceLockedAndRollUp|TestDeclarationDispositionAccountsForThePinnedSourceAndLedger|TestEveryTypedZoomActionHasReverseETLCommandAndCandidate|TestZoomTransportDeclaresTheExecutableSourceAndTypedDestination|TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestRunnableOperationContractsHaveCommands)$'` | PASS (2026-08-22) — all current source-locked inventory, declaration-disposition, missing-foundation, command/action, transport, candidate, and seven-surface mapping assertions pass. The intentionally Foundation-dependent lifecycle test is excluded. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^(TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestProviderInventoryLedgerIsComplete|TestDeclarationDispositionAccountsForThePinnedSourceAndLedger|TestEveryTypedZoomActionHasReverseETLCommandAndCandidate|TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites|TestRunnableOperationContractsHaveCommands|TestZoomTransportDeclaresTheExecutableSourceAndTypedDestination)$'` | Historical PASS before the current Meeting DELETE optional-query regression; it must be rerun after Foundation reaches `main`. It is not current lifecycle evidence. |
| `go test -timeout 20m ./internal/connectors/defs/zoom -count=1` | Historical PASS — 1,748 source contracts, 311 delete contracts, 712 runnable commands, and the exhaustive no-credential preflight sweep. It must be rerun for this surface. |
| `go test -timeout 20m ./internal/connectors/defs/zoom -run '^TestEveryTypedZoomActionHasReverseETLCommandAndCandidate$' -count=1` | Historical PASS — 204 implemented reverse-ETL commands, typed write actions, and generated mutation candidates are an exact one-to-one set. The #4304 continuation adds the closed destination declaration. |
| `go test -timeout 20m ./internal/connectors/engine -run '^TestEveryShippedWriteActionHasExpectedBatchability$' -count=1` | PASS — Zoom actions use the repository default batchability policy while retaining reverse-ETL approval. |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1` | PASS — regenerated root-help transcripts for the declared Zoom surface. |
| `go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1` | PASS — regenerated transcripts are asserted without update mode. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json` | PASS (2026-08-22) — no Zoom definition findings. |
| `go run ./cmd/connectorgen certification-matrix . --connector zoom --check` | PASS (2026-08-23) — regenerated after removing the destination; zero certified capability claims. |
| `go run ./cmd/connectorgen certification-sweep . --connector zoom --check` | PASS (2026-08-23) — 717 rows and 714 CLI commands. |
| `go run ./cmd/connectorgen certification-candidates . --connector zoom --check` | PASS (2026-08-23) — regenerated direct-CLI mutation candidates; no destination capability is claimed. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json` | Expected BLOCKED (2026-08-23) — the only finding is `sources/zoom-operation-source-lock.json`: `parse source lock: json: unknown field "retrieval"`. #4331 owns the rendered-reference v3 source-lock document kind; no Zoom provenance is rewritten while it is pending. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom` | NOT COMPLETED (2026-08-23) — stopped before a result when the full package process became CPU-bound on the shared host. This is not a passing claim; the focused Zoom contract tests above passed and the complete package test will be rerun after #4331. |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable$'` | Expected RED — `zoom_meetings_meetingdelete` fails before I/O because the shared write-query resolver refuses its missing optional `record.occurrence_id`; downstream evidence for #4305. |
| `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestZoomMeetingDeleteOptionalQueryRehearsal$'` in isolated detached SHA `c3f83cbf6eabbae00219566fb02719ca2d6c480d` | PASS — exact Zoom Meeting DELETE query declaration sends the present optional field and omits both absent optional record fields on a fixture-approved loopback request. SHA-bound behavior only; no credentials, Zoom call, certification, or connector-branch ancestry. See [`FOUNDATION-REHEARSAL.md`](FOUNDATION-REHEARSAL.md). |
| `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run '^TestMissingFoundationGapRowsAreSourceLockedAndRollUp$'` | Historical PASS before the delete-only reconciliation — 12 deduplicated gap IDs. The current focused pass above proves 11 catalog gaps, including the eight-row delete-semantics fan-out, with the same 1,329 source-locked rows and 1,299 affected provider operations. |
| `go run ./cmd/connectorgen surface-sync --check` | PASS (2026-08-22) — zero generated field drift. |
| `make connector-boundary` | PASS (2026-08-22) — whole-tree outcome `clean`; 294 files and 552 connector bundles checked; only the repository's six pre-existing, expiring non-Zoom exceptions were reported. |

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
`internal/app/issue_label_warehouse_transport.go:944`. The connector-local matrix records one
live-tested `operation:rest_read` cell but zero certified cells because `fixture_tested=false`.

## Live certification evidence

| Candidate/stage | Result | Honest conclusion |
| --- | --- | --- |
| `api users user` bounded authenticated direct read | PASS — two HTTP 200 exchanges, imported as fingerprint-only `observed_operations` evidence | Implemented with live proof, **not certified**: `operation-specific-fixture-evidence-projection` means no exact fixture can be projected to the operation cell. |
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
