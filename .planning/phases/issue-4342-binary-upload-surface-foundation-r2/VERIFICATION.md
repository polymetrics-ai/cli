# Verification — issue #4342 binary upload CLI and certification foundation

## Planned gates

## Review-remediation gates

- [x] Red/green behavioral coverage for F-4343-01 through F-4343-05, F-4343-07, and F-4343-08 as
  listed in `PLAN.md`; these replace the original overstated lifecycle claims.
- [x] Separately authorized F-4343-06 public GitHub upload-host proof: plan, persisted preview,
  approval, exact `201`, byte/digest read-back, oversize/arbitrary-media/missing-file refusals,
  and an empty retained draft after asset cleanup.

### Remediation execution evidence

- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubReleaseAssetUpload_(InstalledCommandSendsExactBytes|RejectsEnterpriseCrossOriginBeforeIO|RequiresCreatedResponse|EnforcesDeclaredMediaPolicy|RejectsMissingChangedUnsafeOrOversizeFile|EmptyFileAndTerminalFailures)$' -count=1` — pass (3.452s). The double arms both public/Enterprise endpoints and asserts zero calls/authorization headers on cross-origin refusal; it also retains `200`/`202`/`204` receipts as failed outcomes and rejects an unenforceable media list before I/O.
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner -run '^(TestBuildWriteCommandPlansOnlyDeclaredBinaryUploadActions|TestGitHubReleaseAssetUploadBuildsBoundBinaryPlan)$' -count=1` — pass (4.075s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/app -run '^TestBinaryUploadConnectorCommandPersistsPreviewBeforeApproval$' -count=1` — pass (3.916s). This is the App public connector-command lifecycle: root-relative source, no path in state/output, no plan token, persisted preview token, changed-file zero-I/O refusal, exact bytes/SHA-256, and retained `201` receipt.
- [x] Deliberate break: removing `binary_upload` from `planRequiresPersistedPreview` made that App test fail (3.229s); restored before all other checks.
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/bundleregistry -run '^TestGitHubReleaseUploadGuidanceKeepsTheCompositeAliasHonest$' -count=1` — pass (3.618s) after generated-doc refresh.

### Authorized real-provider proof

- [x] The `github-live-upload-proof` encrypted credential was used only in the disposable
  `karthik-sivadas/pm-binary-upload-testbed` draft-release fixture. The public binary built from
  this worktree completed plan → persisted preview → approval-on-stdin → exact 32-byte upload to
  `uploads.github.com`; its retained provider response is `201`. It then refused a 64 MiB + 1 byte
  source, arbitrary `--content-type image/png`, and a missing source before any provider receipt.
- [x] `gh-axi release download` read the named asset back independently; `cmp`, byte count, and
  SHA-256 matched. The proof asset was then deleted through `gh-axi api DELETE`, and the fresh
  retained draft reports `assets: []` for audit. The complete non-secret record is `LIVE-PROOF.md`.
- [x] The generic certification stage is intentionally still `not_live`: it has no safe input
  contract to carry an upload transfer/read-back/cleanup proof, so this observed proof cannot
  falsely promote a generated matrix cell.

- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine` — pass (11.533s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner` — pass (39.064s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/certify` — pass (14.284s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` — pass after the final lint refactor (227.914s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/cli -run '^TestSkillsGenerateMatchesTrackedSkills$' -count=1` — pass (7.631s) after `GOFLAGS='-p=3' go run ./cmd/pm skills generate --dir docs/skills` regenerated the affected GitHub skill.
- [x] `GOFLAGS='-p=3' go run ./cmd/connectorgen validate internal/connectors/defs --json` — 552 connectors, 0 findings.
- [x] `GOFLAGS='-p=3' go run ./cmd/connectorgen surface-sync --check`, `operation-evidence --check`, `certification-candidates --connector github --check`, and `certification-sweep --connector github --check` — all current.
- [x] Generator/docs parity checks from `make verify`: `make docs-check`, `make connectorgen-certification-subject`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, `make connectorgen-certification-sweep`, `make github-parity-artifacts-check`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check` — pass.
- [x] `GOFLAGS='-p=3' go vet ./...`, `GOFLAGS='-p=3' go build ./cmd/pm`, `make tidy-check`, `make lint`, `make smoke-no-build`, and `make agent-contract-check` — pass.
- [x] GSD `verify-work` and `code-review` prompts generated and executed inline; no gap phase was needed. Manual review checked declaration-only source admission, non-executable `file_upload`, evidence classification by intent (not operation name), no direct upload dispatch, and the no-false-pass certification behavior.

## Non-green whole-suite result

- `GOFLAGS='-p=3' go test -timeout 20m ./internal/cli` was attempted and ran for 484.491s. It initially identified generated-skill drift, fixed by the documented skills generation above, but the package still exits non-zero because its broader runtime tests repeatedly dial unavailable local Redis endpoints `127.0.0.1:1` and `127.0.0.1:2`. No task code is changed to suppress those integration failures.
- `GOFLAGS='-p=3' go test -timeout 20m ./...` and aggregate `make verify` were deliberately not invoked as single commands: repository guidance says their 550+ connector suite routinely exceeds the agent command window on this memory-bound machine. All individual non-test gates and every changed-package suite ran; CI retains the aggregate suite.

## Main merge recheck

- Merged current `origin/main` cleanly before publication, then reran `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine`, `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner ./internal/connectors/certify`, and `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` — all passed. Regenerated `operation-evidence --write-fixed-100` for the merged declaration fingerprint; its check and the certification subject/matrix/candidates/sweep checks then passed.

## Post-PR website correction

- GitHub's `Website checks` and `Website generated data` initially failed because this intent changed three generated website artifacts. Ran `cd website && pnpm run gen:website-data`, then verified `pnpm run lint` (warnings only, exit 0), `pnpm run typecheck`, `pnpm run test:unit` (80 tests), `pnpm run test:scripts` (34 tests), and `pnpm run build` — all passed. The build emits existing Better Auth default-secret warnings while statically collecting pages but exits 0.
- When `main` advanced again during CI readiness, merged it into the published branch and reran `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` (pass, 336.788s) plus `cd website && pnpm run gen:website-data` (no resulting diff).

## Constraint

The authorized provider proof is deliberately bounded to one disposable draft release and is not
reused as a generic credential or matrix claim. A missing stage-specific transfer/read-back/cleanup
contract remains `not_live`, never a passing transfer assertion.
