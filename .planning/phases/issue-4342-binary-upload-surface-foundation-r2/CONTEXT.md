# Context — issue #4342 binary upload surface foundation

## Task Delivery Header

- Issue: Refs #4342 — feat(connectors): add binary upload CLI and certification surface
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, with committed implementation, targeted behavioral tests, generated-file checks, and applicable local verification recorded.
- Working branch: fm/cli-binary-upload-surface-foundation-r2
- Task: Expose only declaration-bound binary upload write actions through a new `binary_upload` connector-command intent, preserving the existing plan → preview → approval → execute pipeline; add a separate truthful upload certification projection and evidence stage.
- Verification: Targeted engine, commandrunner, CLI, certification, and connectorgen tests; `connectorgen validate`, `surface-sync --check`, operation-evidence checks, generated docs checks, and applicable package-level Go checks with `GOFLAGS=-p=3`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A binary-upload command is admitted only for a declared binary/base64/multipart file write action | live | Commandrunner preflight accepts the declared GitHub binary action and rejects ordinary JSON, non-file multipart, undeclared, and mismatched source-field actions before any request runner can execute. |
| The command follows the established approval-bound write path | live | `TestBinaryUploadConnectorCommandPersistsPreviewBeforeApproval` drives App plan → persisted preview → approval → exact-byte provider-double execution, including no-token-before-preview and changed-file zero-I/O refusal. A separately authorized GitHub upload-host proof is recorded in `LIVE-PROOF.md`. |
| Upload certification cannot call a refusal a pass | live | Certification stage test makes a declared candidate return a refusal and asserts capability/stage status is `blocked` or `not_live`, `Passed=false`. There is no checked-in transfer/pass branch; any future pass requires separately evidenced bytes, digest, response, read-back, and cleanup. |
| Upload is independently represented in generated certification/evidence surfaces | live | Sweep and operation-evidence tests assert `binary_upload` has its own class and rejects `file_upload` as executable. |
| Help, manuals, and website projections describe the same closed surface | live | The project-root path contract is executed by the App test and named by command help. The legacy composite alias describes only its remaining multi-file/clobber/label gap, and generated manual, skill, and website checks share that source. |

## Decisions recorded by discuss-phase (inline fallback)

- Captain ruling fixes `binary_upload` as a release-required, eighth first-class capability; no product question remains open.
- A `binary_upload` command is a write-action binding, never an operation executor or raw upload primitive. It must carry `write`, not arbitrary body/path/URL input.
- The command reuses `BuildWriteCommand` and the existing connector-command plan. It is therefore never executable directly by `commandrunner.Run`.
- `file_upload` remains declarable but unexecutable; it must not be mapped to `binary_upload` or receive an executor.
- A successful safety refusal is still useful evidence, but reports `blocked` or `not_live` and the owning stage is not passing. Only an evidenced transfer plus cleanup can set upload `pass`.
- Existing binary upload/base64/multipart write paths retain source confinement, digest binding, and byte caps. Raw/base64 allow-lists are now executable policy: a declaration must admit the executor's fixed `application/octet-stream` media type or is refused before I/O.

## Frozen review remediation finding set — 2026-08-24

- **F-4343-01 (critical):** a GitHub Enterprise credential can cross to public
  `uploads.github.com` through the fixed upload base URL.
- **F-4343-02 (high):** `file_path` is persisted and printed by the plan/preview path.
- **F-4343-03 (high):** the GitHub upload reports any `2xx`, rather than only declared `201`, as success.
- **F-4343-04 (medium):** raw/base64 media allow-lists promote the public surface but are not enforced at execution.
- **F-4343-05 (high):** the original lifecycle/certification evidence overstated nonexistent end-to-end tests.
- **F-4343-06 (high):** no evidence yet crosses GitHub's actual upload-host boundary.
- **F-4343-07 (low):** generated legacy guidance denies an executor used by an implemented sibling.
- **F-4343-08 (high):** `--file-path` says “Project-relative path to the release asset bytes,”
  but planning roots the path at `<project>/.polymetrics`. Both natural user combinations were
  observed to fail before planning/provider I/O: project-root file plus `live-upload-proof.bin`
  resolves to `<project>/.polymetrics/live-upload-proof.bin`; moving the file there and passing
  `.polymetrics/live-upload-proof.bin` resolves to a double-prefixed
  `<project>/.polymetrics/.polymetrics/live-upload-proof.bin`. The sole currently working
  combination (file beneath `.polymetrics`, bare filename) is undiscoverable from public help,
  manual, skill, or website output. Remediation must make project-root relative paths work and
  must never require callers to know `.polymetrics`. This defect is inherited from `origin/main`:
  main already ships the identical file-path summary on this command. The branch changes that
  command from `intent=reverse_etl` with “preview is optional” to `intent=binary_upload` with
  preview required, so the user-visible lifecycle change must be disclosed and behaviorally proved.
- **F-4343-09 (high):** the human preview for binary-upload plan
  `rplan_67fb5cbee429ffa7` emitted exactly:
  `Reverse plan rplan_67fb5cbee429ffa7 previews releases assets upload via releases_release_id_assets2`;
  `- releases_release_id_assets2 executes a live mutation only after approval; dry run performs no external call`;
  `- resolved request: POST https://uploads.github.com/repos/karthik-sivadas/pm-binary-upload-testbed/releases/375528670/assets?name=live-upload-proof.bin`.
  It emitted no `Approval token:` line. The shorthand renderer is
  `internal/cli/cli.go:1897-1914`; it calls `PreviewConnectorCommandPlan` and prints a token only
  when the returned plan contains one. `pm reverse preview` has a different renderer at
  `internal/cli/cli.go:2177-2202`, but `PreviewReversePlan` delegates connector plans straight back
  to that same app method (`internal/app/app.go:2481-2482`), so neither entry point can recover a
  token the stored plan deliberately omits. For a generic action with no typed confirmation,
  `PreviewConnectorCommandPlan` only dry-runs and returns its original planned state
  (`internal/app/app.go:2421-2435`); it neither persists preview evidence nor mints a token.

  The plan renderer had already exposed the pre-minted token (`internal/cli/cli.go:1835-1840`), and
  `RunReverseETL` only requires a persisted preview for an operation, issue-label transport, or a
  typed confirmation (`internal/app/app.go:2796-2797,2869-2870`). Therefore the upload is
  executable with that pre-preview token, but the branch's required plan → preview → approval
  contract is bypassable. The declaration sweep found 2,192 implemented generic write commands:
  2,191 `reverse_etl` and this one `binary_upload`, all with no operation; 2,162 have no typed
  confirmation and share the pre-minted-token path, while 30 destructive generic commands persist
  their preview and issue the post-preview token. The 283 implemented `direct_write` commands use
  the operation-persisted path and are not affected. This generic behavior exists on `origin/main`
  (2,191 reverse-ETL commands, including 2,161 no-confirmation commands); main documents preview
  as optional, and its users can execute via the human plan token. The remediation in this PR must
  make binary-upload plans emit no token until a persisted preview, then issue one bounded
  human-only token from that preview, without silently changing the legacy reverse-ETL contract.

## Review remediation disposition — 2026-08-24

- **F-4343-01:** closed by `allowed_base_url_origins` and the dual-server zero-I/O regression: an
  Enterprise/API origin cannot send its credential to the fixed public upload origin.
- **F-4343-02:** closed by declaring `file_path` redacted and by the App lifecycle test's state and
  safe-output sentinel assertions.
- **F-4343-03:** closed by declaration-owned `success_statuses:[201]`; every other 2xx becomes a
  retained failed provider receipt.
- **F-4343-04:** closed by runtime media-policy enforcement for raw/base64 upload executors.
- **F-4343-05:** closed by the real App lifecycle test and its deliberate persisted-preview-gate
  break; prior evidence claims were rewritten rather than reused.
- **F-4343-06:** `LIVE-PROOF.md` records fresh actual GitHub-host transfer, exact byte/digest
  read-back, oversize/arbitrary-media/missing-file refusals, and an empty audited draft cleanup.
  The original disposable release was deleted during its first cleanup, which made it unauditable;
  the fresh draft is deliberately retained empty. The generated stage remains non-passing until it
  owns this full proof contract.
- **F-4343-07:** closed by the rendered-guide regression and regenerated manual/skill/website rows.
- **F-4343-08:** closed by resolving `binary_upload` source files against the project root, while
  retaining root confinement; no caller names `.polymetrics`.
- **F-4343-09:** closed only for `binary_upload`: plan emits no token, persisted preview mints the
  one-time token, and execution requires that preview. The inherited generic no-confirmation
  `reverse_etl` behavior remains deliberately out of this PR's scope and is disclosed in its body.

## GSD fallback

`scripts/gsd prompt discuss-phase issue-4342-binary-upload-surface-foundation-r2` and `scripts/gsd prompt plan-phase issue-4342-binary-upload-surface-foundation-r2 --tdd` were resolved. The compatible isolated Pi/GSD runtime is unavailable in this Codex worktree and the repo's single-worker contract forbids role spawning, so the generated workflow is being performed inline. This records the same decisions, TDD slices, verification, and later review evidence without weakening any gate.

## Required skills loaded

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `vercel-react-best-practices`, and `vercel-composition-patterns`.

The website skills are loaded for the generated documentation projection; no website component architecture or dependency change is intended.
