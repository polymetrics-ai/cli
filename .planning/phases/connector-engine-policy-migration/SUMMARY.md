# Summary — connector engine/direct-read policy migration

## Completed slice

- Replaced provider-named lower-bound formatting with generic `param_format: "rfc3339_utc"` plus bounded `operator_prefix` (`>=`, `>`, `<=`, `<`).
- Replaced provider-named repository contents direct-read output policies with generic:
  - `repository_contents_file_metadata`
  - `repository_contents_directory`
- Updated GitHub CLI surface metadata to use the generic repository contents policies.
- Preserved security behavior: repository path sensitivity checks still run before network access; file metadata responses still redact `content` and `download_url`; directory policy still rejects file-shaped responses.
- Added/reworked regression fixtures so generic comparison-prefix formatting and repository contents redaction are proven by synthetic non-provider connector shapes.
- Updated validator/schema/conformance/docs contracts.
- Removed 12 drained connector-boundary exceptions; boundary applied exceptions dropped from 24 to 12 with 0 findings and 0 warnings.

## GSD / TDD evidence

- GSD adapter health: `scripts/gsd doctor`, `scripts/gsd list`.
- Manual universal-loop fallback recorded because `scripts/gsd prompt programming-loop --help` is not registered in this checkout.
- Red evidence: before production code, updated engine tests failed to compile on missing `IncrementalSpec.OperatorPrefix`.
- Green evidence: targeted and full gates pass; see `VERIFICATION.md`.

## Safety

No secrets, no live connector checks, no new dependencies, no generic write tools, and no reverse-ETL behavior changes beyond existing smoke verification.
