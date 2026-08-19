# GSD discuss-phase #4087 (inline/manual fallback)

## Inputs reviewed

- GitHub issue #4087 and parent #4015.
- `AGENTS.md`, the required-skill routing, GSD Pi adapter, issue-agent contract, and CLI parity reference.
- `internal/app/sync_modes.go`, `internal/app/app.go`, `internal/app/transport_dispatch.go`, and current sync/transport tests.

## Resolved questions

| Question | Resolution | Evidence |
| --- | --- | --- |
| Does the reported defect exist? | Yes. Both affected alias families omit `ContractMode` in normal and persisted-legacy parsing. | `sync_modes.go:54-55,58-59,94-95,98-99` |
| How is the legacy branch selected? | A missing contract makes `IsContractMode()` false, so `RunETL` proceeds past `app.go:1104` to catalog and legacy ETL. The base also suppresses typed admission for `LegacyCompatibility`; the corrected aliases must not carry that suppression. | `sync_modes.go:142-144`, `app.go:1093-1124` |
| Which canonical modes represent the aliases? | Full overwrite and incremental dedupe, respectively. | Closed vocabulary in `internal/synccontract/mode.go`; #3810 contract |
| Is connector-specific code needed? | No. The parser is application-level and must remain connector-neutral. | #4087 portability rule |
| Do help/docs/generated surfaces need content changes? | Yes. The public names stay unchanged, but their prior descriptions falsely promised legacy dedupe execution. Runtime help, generated CLI docs, and website docs now describe typed admission and pre-I/O refusal until a transport is admitted. | #4087 acceptance criterion 3 |

No product, destructive, dependency, connector-specific, credential, or reverse-ETL decision is required.
