# Discussion log — issue #4342

## Inputs reviewed

- `AGENTS.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-all-batch-surface-gap-map-r1/report.md`
- `docs/migration/conventions.md` binary and multipart contracts
- Existing engine binary-upload preparation/execution and GitHub release-asset declaration
- Existing certificate binary-download stage, report shape, sweep, and operation-evidence projection

## Resolved implementation boundaries

1. The new public intent is distinct from `direct_write`, `reverse_etl`, `binary_download`, and `file_upload`.
2. Command input remains a closed list of declared record flags. A file path is accepted only through the write action's declared source field and project-root checks already enforced by the engine.
3. Upload commands use the existing plan, preview, approval, execute machinery; plain dispatch must report `ErrNotWriteCommand` rather than sending bytes.
4. Certification separates download and upload in the report and sweep. A `blocked`/`not_live` upload stage has `Passed=false`.
5. The initial real binding is GitHub's existing `releases_release_id_assets2` binary write action, changing its already executable connector path from `reverse_etl` to `binary_upload`; generated manual/skill/website material follows the definition.
6. Certification candidates require declaration-owned setup, upload, read-back, and cleanup inputs. No credentials or raw config values are included in artifacts.

## TDD sabotage criterion

After the happy-path test goes green, temporarily remove the binary-upload intent admission from commandrunner and run the targeted test. It must fail because the plan cannot be created. Restore the production code, rerun the test, and record both outcomes in the ledger.
