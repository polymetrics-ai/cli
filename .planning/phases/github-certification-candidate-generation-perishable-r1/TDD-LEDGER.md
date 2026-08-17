# TDD ledger — GitHub certification candidate generation

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Surface candidate projection | **Red:** `go test -timeout 20m ./cmd/connectorgen -run 'TestBuildGeneratedReadCandidates'` failed at compile time because `buildGeneratedReadCandidates`, the generation specification, and connector-owned cohort type did not exist. | **Green:** after the generic projection and schema support, the same focused command passed. It derives command tokens, required flag values, `--credential`, `--json`, and `/response` `object_or_array` assertion, while excluding writes and unavailable commands. | Candidate data must be source-derived and assertion-bearing; it is never a pass. |
| Perishable classification | **Red:** an engine schema `minimum` keyword was rejected by the repository's deliberately limited schema compiler. | **Green:** literal connector-owned cohorts validate their `command_count`, contain 50/50/46/46 commands (192 total), and generate 31/23/21/22 direct reads (97 total) without a shared GitHub identifier. | The remaining 19/27/25/24 reverse-ETL commands (95 total) are explicitly deferred to the mutation-fixture lifecycle. Sweep status arithmetic remains 1,571. |
| Live run and failure proof | **Red:** regenerating the known passing `copilot configuration view` case with `/response` declared as `string` made its own certification stage fail: `declared output at /response has the wrong type`. | **Green:** restore `object_or_array`, regenerate, and rerun the same named stage; it passes. The four serial cohort runs executed all 97 generated reads: 34 produced-value passes and 63 non-passes. | `LIVE-RESULTS.md` contains the bounded outcome accounting and named product defects. Credentials were environment-only and no raw success values were persisted. |

## Scope boundary

The existing direct-read runner rejects mutation contracts by construction.
That is the named, intentional boundary of this delivery, not a test exemption:
the generator never produces a pass for an unexecuted operation and does not
invent mutation fixtures, plan/preview/approval, read-back, or cleanup.
