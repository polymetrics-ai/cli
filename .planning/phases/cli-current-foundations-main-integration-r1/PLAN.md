# Current Foundations Main Integration r1

## Task Delivery Header

- Issue: Refs #4302 — status/export foundation; Refs #4303 — declarative reverse ETL; Refs #4305 — structured REST body; Refs #4306 — source import; Refs #4307 — closed operation runtime.
- Base branch: `main` at `e62ae21d428f0d27225f9bff564dc2cd797f6b65`.
- Merges into: `fm/cli-current-foundations-main-integration-r1 → main`.
- Delivery: One human-gated rollup PR against `main`, created only by the later no-mistakes stage after this branch has the exact qualified component heads, local production gates, actual-provider qualification, and a complete evidence manifest.
- Working branch: `fm/cli-current-foundations-main-integration-r1`.
- Task: Preserve the exact five component histories and compose the source-import, closed-runtime, status/export, structured-body, and declarative reverse-ETL foundations without broadening API authority or dropping generated, provider-owned, safety, App, CLI, documentation, or result surfaces.
- Verification: Targeted red-green tests for every conflict correction; focused engine, commandrunner, App, CLI, sync-transport, source-import, generator, and regression tests; `go vet ./...`; `go build ./cmd/pm`; `connectorgen validate`; `surface-sync --check`; generated CLI/help/manual/website checks; `connector-boundary`; `make verify`; and the specified real-provider qualifications.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Exact component heads are immutable and approved before composition | live | `data/cli-current-foundations-main-integration-r1/input-manifest.json` records a matching remote pull ref and Firstmate qualification for every merged head. |
| No component contract is lost in conflict resolution | live | Production-shaped red tests exercise all cross-feature routes and remain green from the final rollup SHA. |
| Installed CLI and persisted App preserve declared operations and provider results | live | Installed-binary and App tests assert reachable declared surfaces, plan/apply/ack behaviour, result fields, and generated help. |
| Actual-provider claims are not synthetic-only | live | Qualification report records sanitized real-provider status/export, declared write, source-import, and reverse-ETL assertions from the integrated SHA. |

## Scope and safety invariants

- Preserve component commits with merge commits; do not cherry-pick, rewrite, reset, force-push, close, retarget, or merge any component PR.
- Every integration or conflict-resolution commit includes `Refs #4302`, `Refs #4303`, `Refs #4305`, `Refs #4306`, and `Refs #4307`.
- No connector-name branch, generic HTTP method/path/header/body/action authority, generic shell, or generic SQL write capability is permitted.
- Declarations own operation identifiers, path/query/header/body mappings, bounded transfer rules, status metadata, and all non-secret provider result fields. Credentials stay in existing encrypted/masked boundaries and never enter evidence.
- Reverse ETL remains plan → preview → approval → execute → durable acknowledgement.

## Inputs and merge order

1. Confirm the immutable source-import head for #4306 / #4312, then merge it without rewriting its history.
2. Confirm the immutable closed-runtime head for #4307 / #4311, inspect source-import/runtime overlap, add red tests before any correction, and merge it.
3. The exact #4302 / #4308 head has Firstmate's terminal qualification; inspect and merge it after #4306 and #4307 without rewriting its history.
4. Wait for Firstmate's exact published, qualified structured-body head for #4305; do not substitute the visible pre-remediation branch.
5. Wait for Firstmate's exact published, qualified reverse-ETL head for #4303; do not substitute the currently failing #4304 head.
6. Compose each newly qualified exact head in the declared dependency order, resolving only with production-shaped red → green → refactor evidence.

## TDD integration slices

| Slice | Red | Green | Refactor / regression guard |
| --- | --- | --- | --- |
| Declaration-bound request shaping | A real declaration combining structured body, typed header, exact query, and exact path must fail before I/O for malformed, unknown, oversized, duplicate, CR/LF, or cross-bound values. | The request materializes only declaration-owned fields and reaches the declared operation exactly. | Re-run scalar, form, SCIM, binary, and specialized GitHub cases unchanged. |
| Bounded response/status composition | A terminal status-only 4xx/5xx after normal retry handling must retain final metadata, while binary/text GET errors must remain errors and produce no output file. | Closed status/text/binary operations preserve typed status, headers, body bytes, and bounds. | Re-run loader, output, retry, and download regressions. |
| Source declaration reachability | An accepted lock-verified source operation must fail if it cannot reach generated command/help surfaces. | Exact source bytes/count/SHA import produces the fixed declaration and installed surface. | Re-run lossless/oversized/malformed source and generator validation cases. |
| Typed reverse-ETL composition | Multiple named actions must fail if plan/apply/ack or any provider result field is not persisted through App and installed CLI. | Independently selectable actions complete plan, preview, approval, apply, durable acknowledgement, and provider-result preservation. | Re-run existing typed destinations and direct-write bindings. |

## GSD and skills record

- GSD adapter health and command resolution: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`; `go run ./cmd/agentcontractgen check`.
- Inline/manual fallback: the task's canonical single-worker contract and this runtime forbid spawning the generated GSD roles, so the generated prompts are executed and recorded inline.
- Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-documentation`.
- CLI parity is mandatory: runtime help, bare namespace behavior, docs/cli, website, generated help/manual/discovery surfaces, JSON/stdout-stderr contract, and reverse-ETL safety text are verified together.

## Checkpoints

- [x] Isolated worktree, clean `main` base, no-mistakes daemon health, GSD adapter, skills, base commit, and published-ref intake verified.
- [ ] All five Firstmate-qualified immutable component heads received and re-read through `gh-axi` (three currently eligible: #4306, #4307, and #4302).
- [ ] Component history merged in dependency order with conflict probes and red-green-refactor evidence.
- [ ] Full local and provider qualifications pass from one final integrated SHA.
- [ ] Evidence manifest/report is complete, temporary qualification material is removed recoverably, branch is clean, and implementation is committed.
