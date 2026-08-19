# UAT — #4081 warehouse-mediated Transport demo

**Mode:** inline/manual `verify-work` fallback. The official prompt was
resolved with `scripts/gsd prompt verify-work
issue-4081-warehouse-mediated-transport-demo-r1 --auto`; this issue phase has
no compatible numeric/GSD-role execution route, and the repository contract
forbids spawning one.

| Deliverable | Automated acceptance evidence | Result |
| --- | --- | --- |
| D1 — fail-closed production construction and durable reopen | `TestOpenInstallsGitHubWarehouseMediatedTransport`, `TestGitHubWarehouseStageReopensDurableReceiptAfterSourceReferencesAreDiscarded`, fresh-open isolation, and manifest/WAL/Parquet corruption cases in `internal/app/transport_construction_red_test.go` | Pass |
| D2 — approval-bound exact GitHub route | `TestGitHubIssueLabelDestinationRejectsUnapprovedOrMismatchedOrExpiredOrReplayedPlanBeforeProviderWrite`, cleanup tamper/404/replay tests in `internal/app/github_warehouse_transport_approval_test.go`, and `TestRunETLRejectsDestinationApprovalOutsideClosedTransportRoute` | Pass, including race run |
| D3 — closed operator carrier and no approval leakage | `internal/cli/etl_transport_test.go`: exact grammar, project-free manuals, one 4 KiB-bounded stdin line, forbidden carrier fields, legacy runtime preservation, and safe preview DTOs | Pass |
| D4 — real built-binary walking slice | `TestPMBinaryExecutesGitHubWarehouseTransportLifecycle`: fresh binary, source page, WAL/DuckDB/Parquet/manifest, reopened singleton, typed POST, independent GET read-back, acknowledgement/checkpoint, 204 cleanup, separately approved 404 cleanup, replay refusal, zero residue | Pass |

No human judgment is required for D1–D4: each is a deterministic local
contract with an explicit passing test. The real provider is deliberately not
claimed as certified: the normal approved GitHub App/disposable
`Polymetrics-Cert` boundary was not supplied to this isolated task project, so
the safe result remains `LIVE_PROVIDER_BLOCKED_NO_APPROVED_GITHUB_APP_BOUNDARY`
before any real provider I/O.
