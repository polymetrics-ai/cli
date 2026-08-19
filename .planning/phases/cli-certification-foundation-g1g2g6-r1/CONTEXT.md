# Context: connector certification foundation G1/G2/G6

The Firstmate launch brief, the certification-fixture scout report, and the existing certification design determine this foundation slice. The task runs non-interactively: the captain already locked the parity taxonomy, evidence authority, exclusions, delivery base, and proof target. The manual discussion conclusion is therefore to use those supplied decisions without creating a second certification authority.

Locked decisions:

- The generated projection has exactly eight engine kinds (`rest_read`, `rest_write`, `etl`, `reverse_etl`, `binary_download`, `file_upload`, `cdc`, `changefeed`) and five parity classes (`direct_read`, `direct_write`, `etl`, `reverse_etl`, `binary`).
- A write action remains a `direct_write` class; its generated `write_action` retains `create`, `update`, `upsert`, `delete`, or `custom` so delete is independently selectable later.
- A missing or inconsistent projection is a generator error, never a silent N/A. Native managed-destination database transport is a real `reverse_etl` route despite a generic direct `write:false` capability.
- Accepted evidence remains under `internal/connectors/certifications/`; multi-record publication validates the entire batch before publishing any record and each immutable file is atomic/no-replace.
- The live script must stage a draft, import it, generate the connector shard, then check; a global `--all` refresh/fan-in remains out of scope.

## Manual GSD fallback

`scripts/gsd doctor`, all lifecycle `sources` calls, and generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved before work. This runtime cannot provide compatible isolated GSD worker worktrees and the captain prohibits role spawning, so the generated workflows are performed inline. This is a documented manual fallback, not a lifecycle waiver.
