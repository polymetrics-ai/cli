# Plan: Gong CLI Parity Parent Orchestration

Parent issue: #133
Parent branch: `feat/133-gong-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/232 (draft, base `main`)
Default branch: `main`

## GSD command path

- `scripts/gsd doctor` — pass.
- `scripts/gsd verify-pi` — pass.
- `scripts/gsd list --json` — pass; output truncated by harness but command exited 0.
- `scripts/gsd prompt plan-phase 133 --skip-research --tdd` — rendered plan prompt to `/tmp/gsd-plan-phase-133.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-cli-parity --dry-run` — adapter returned `unknown GSD command: programming-loop`.
- Fallback: use repo-local Pi `/pm-orchestrate` + `/pm-gsd-loop` prompt bodies and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; record manual-GSD fallback until `scripts/gsd` exposes `programming-loop`.

## Required skills loaded

- GSD/runtime: `gsd-core`, `caveman`.
- Go/CLI/connector: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-lint`.
- Repo references: `required-skills-routing.md`, `gsd-pi-adapter.md`, `cli-help-docs-website-parity.md`, connector migration handoff/conventions/design docs.

## Mission

Deliver Gong connector CLI parity against the official public Gong OpenAPI 3.0.1 spec at `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=` and parent issue #133.

Official spec fetch evidence on 2026-07-09:

| Metric | Count |
|---|---:|
| Paths | 57 |
| Operations | 67 |
| GET | 28 |
| POST | 27 |
| PUT | 8 |
| PATCH | 1 |
| DELETE | 3 |

Current bundle baseline:

- `internal/connectors/defs/gong/` exists.
- Streams: `users`, `calls`, `scorecards`.
- `api_surface.json`: 10 entries, 3 `covered_by.stream`, 7 legacy `excluded` rows.
- The exact public spec shows `/v2/calls/extensive` and `/v2/calls/transcript` are `POST` read-query operations, not `GET` out-of-scope exclusions.

## Sub-issue queue and dependencies

| Issue | Lane | Write scope | Dependencies | Initial state |
|---:|---|---|---|---|
| #144 | Operation ledger | `internal/connectors/defs/gong/api_surface.json`, docs, connector-specific validation tests | none; unlocks other lanes | local critical path |
| #141 | CLI surface metadata | `internal/connectors/defs/gong/cli_surface.json`, metadata validation/tests | #144 for exact operation IDs/classification | queued |
| #142 | Help renderer/docs | help/docs/website/generated artifacts | #141 metadata | queued |
| #143 | Stream runner | Gong stream definitions/schemas/fixtures/runner tests | #144, may parallel with #145 after ledger | queued |
| #145 | Direct reads | direct-read operation metadata/executor tests | #144; may need #146 body policy for POST query shapes | queued |
| #146 | Advanced body/binary | bounded body/binary policy and fixtures | #144 | queued |
| #147 | Sensitive/admin policy | `writes.json`, sensitive policy/redaction/confirmation tests | #144 and #146 for payload policy | queued |

## Orchestration decision

Decision: `not_spawned_runtime_capability_missing`.

Reason: this Pi API session exposes only `read`, `bash`, `edit`, and `write`; no `subagent` tool is available, so mutating workers cannot be spawned into isolated worktrees from this session. Continue with #144 as `local_critical_path` and record future worker prompts/handoffs in the state ledger.

## Parent PR plan

1. Create parent planning artifacts and #144 red test.
2. Land green #144 slice locally.
3. Commit only scoped files; do not add untracked `PI_CONNECTOR_PROMPT.md` unless explicitly requested.
4. Push `feat/133-gong-cli-parity` after local targeted validation.
5. Keep draft parent PR #232 open until all required sub-issues are integrated and final verification passes.
6. Do not merge parent PR to `main`; final merge is human-gated.

## Full-surface safety constraints

- No secrets requested, printed, stored, summarized, or embedded in fixtures.
- No credentialed Gong checks unless explicitly requested.
- No generic HTTP write, raw GraphQL mutation, generic shell write, or generic SQL write surface.
- Reverse ETL remains plan → preview → approval → execute.
- Sensitive/admin/destructive endpoints are typed reverse-ETL candidates, not permanent exclusions.
- Binary/input payload endpoints require bounded local input/output policy and no path traversal or broad paths.

## Verification plan

Targeted during #144:

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -count=1
go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1
```

Broader before parent handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```


## 2026-07-10 integrated local critical path update

Because this Pi API session still has no subagent tool and the remaining Gong definition write scopes collide, issues #141, #142, #143, #145, #146, and #147 were executed inline as a single local critical-path integration after #144.

Integrated coverage:

- 12 stream-covered GET endpoints (expanded beyond `users`, `calls`, `scorecards`).
- 16 implemented bounded GET direct reads using `json_redacted`.
- 23 typed JSON reverse-ETL write actions in `writes.json`.
- 16 typed operation metadata rows for POST read-query and multipart/top-level-array executor gaps.
- No generic raw HTTP write, arbitrary body, shell write, SQL write, or credentialed Gong check was introduced.

## 2026-07-10 engine-shape implementation expansion

Created follow-up implementation issues from the #146 analysis to close the remaining Gong executor gaps:

- #252 — typed POST read-query operation execution.
- #253 — schema-gated top-level JSON array request bodies.
- #254 — bounded typed multipart upload support.

Runtime decision: no subagent tool is available in this Pi API session and the write scopes overlap in `internal/connectors/engine`, `internal/connectors/commandrunner`, `cmd/connectorgen`, and Gong definitions, so the parent orchestrator records `local_critical_path_runtime_capability_missing` and executes the slices sequentially on `feat/133-gong-cli-parity`.

## 2026-07-22 completion and merge-readiness cycle

User objective: finish remaining Gong surface, make transcript executable, enable connector immediately on merge, and prepare parent PR #232 for human merge.

Execution decision: `not_spawned_runtime_capability_missing` followed by `local_critical_path`. The current Pi harness exposes `read`, `bash`, `edit`, and `write`, but no `subagent` tool. Coupled changes also overlap in CLI dispatch, engine safety, Gong definitions, generated docs, and shared parent artifacts.

Required slices:

1. **CLI help parity (#142)** — make `pm help gong`, `pm gong`, `pm gong --help`, and command-level `--help` render connector metadata without opening a project or requiring credentials.
2. **Approval/upload safety (#254)** — bind approved file uploads to content digest, propagate the approved digest to the upload transport, snapshot and verify the exact bytes before any HTTP request, enforce multipart byte limits while snapshotting/streaming (not only preflight `stat`), and keep paths/content out of rendered plans.
3. **Remaining typed POST reads (#252)** — replace all 10 planned Gong operation blockers with connector-authored typed flags and executable bounded/redacted direct reads. `calls transcript` is mandatory.
4. **Coverage and generated parity (#133)** — classify the 10 operations as executable direct reads, regenerate connector manual/skill and website catalog, and retain reverse ETL plan → preview → approval → execute.
5. **Final readiness** — validate that every documented typed POST example supplies a schema-valid minimum request, run CLI/upload/definition review lanes concurrently, then run full local gates and clean local Codex review coverage on the current head. Clear the newly published GO-2026-5970 CI gate with a minimal existing indirect `golang.org/x/text` upgrade (no new dependency), update PR closing keywords, and preserve the human merge gate. Per user direction, skip CodeRabbit, Claude, and Copilot review requests for this change set.

Parallel completion lanes (parent orchestrator owns integration):

- Lane A: CLI help and response-cap review (`internal/cli/**`, `commandrunner/**`).
- Lane B: upload approval/snapshot review (`internal/app/**`, `connectors.go`, `connsdk/**`, `engine/write*`).
- Lane C: Gong operation schema, minimum-example, and API-ledger review (`defs/gong/**`, Gong generator tests).
- Lane D: generated docs/website refresh after Lane C.
- Lane E: targeted checks in parallel, followed by one repository-wide verification barrier and local Codex review.
- Runtime note: the current Pi harness still has no `subagent` tool, so independent read/test lanes run concurrently through parallel tool calls while production file ownership remains serialized to avoid collisions.

TDD order:

- Red CLI tests for dynamic connector help and bare namespace behavior.
- Red upload tests proving same-size/same-mtime content changes invalidate approval, approved bytes are re-verified from a private snapshot before network send, and post-validation growth cannot exceed `MaxBytes`.
- Red Gong coverage tests requiring zero planned direct reads, implemented transcript metadata, and schema-valid minimum examples for every newly enabled POST read.
- Capture the pushed-head `govulncheck` red for GO-2026-5970 (`x/text` v0.36.0; fixed v0.39.0), perform only the fixed-version upgrade, then run tidy, verify, tests, vet, and govulncheck.
- Smallest green implementation per slice, then refactor and generated artifact refresh.

No credentialed Gong requests, live writes, new dependencies, raw body flags, generic HTTP writes, or parent merge to `main`.
