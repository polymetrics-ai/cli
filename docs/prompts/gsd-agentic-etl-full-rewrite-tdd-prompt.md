# Prompt: Agentic ETL Full Rewrite With GSD TDD Loop

Use this prompt when asking Codex, Claude, or another implementation agent to build the production-grade Polymetrics Go CLI platform using the GSD universal programming loop and strict red-green-refactor TDD.

```text
Act as an expert Go software engineer, systems architect, and security-minded implementation agent.

Repository:
Polymetrics Go CLI monolith.

Primary goal:
Implement the complete agentic ETL and reverse ETL platform as a Go-only CLI monolith that supports dependency-free local execution and optional runtime-backed execution with PostgreSQL, DragonflyDB, and Temporal. Preserve the connector design ideas from the original Polymetrics/Ruby connector system while adopting the strongest agent-facing CLI patterns from googleworkspace/cli.

Operating mode:
Use the GSD universal programming loop with strict TDD.

GSD settings:
- workflow.use_worktrees=false
- workflow.tdd_mode=true
- Run this as a local-first implementation loop.
- Do not mark the phase complete until PRD coverage, TDD ledger, verification, docs, and run-state artifacts are updated.

Phase name:
agentic-etl-platform

Required context to read before coding:
- .planning/PROJECT.md
- .planning/ROADMAP.md
- .planning/STATE.md
- .planning/config.json
- docs/architecture/repo-profile.json
- POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md
- POLYMETRICS_AGENTIC_ETL_GO_PLAN.md
- README.md
- docs/architecture/runtime-dependencies.md
- docs/runtime/SETUP.md
- docs/prompts/universal-programming-loop-prompts.md
- docs/prompts/gsd-runtime-dependencies-prompt.md
- internal/cli/*
- internal/app/*
- internal/connectors/*
- internal/vault/*
- internal/perf/*
- scripts/runtime.sh
- scripts/setup-runtime-macos.sh
- scripts/setup-runtime-linux.sh
- deploy/compose/polymetrics-runtime.yml

External reference to study:
- If /Users/karthiksivadas/Development/googleworkspace-cli exists, inspect it.
- Otherwise clone https://github.com/googleworkspace/cli into /Users/karthiksivadas/Development/googleworkspace-cli and inspect it locally.
- Study README.md, AGENTS.md, CONTEXT.md, crates/google-workspace-cli/src/main.rs, commands.rs, executor.rs, auth.rs, credential_store.rs, output.rs, generate_skills.rs, and crates/google-workspace/src/validate.rs.

Patterns to adopt from googleworkspace/cli:
- Manifest/schema-driven command, docs, and skills generation.
- Agent-first stable JSON output and stable typed exit codes.
- Incremental pagination and streaming output instead of buffering large runs.
- Credential precedence: env token, credential file, encrypted local credentials, OS keyring or file fallback.
- Strict terminal output sanitization and parse-boundary validation for agent-supplied inputs.
- Helper workflow commands that are distinct from raw connector operations.
- Generated agent skills, recipes, and an agent-facing CONTEXT.md.
- Clear AGENTS.md safety rules for future implementation agents.

Non-negotiable safety rules:
- Never print, log, serialize, snapshot, or commit secret values.
- Never expose generic shell execution to agents.
- Never expose generic external HTTP write tools to agents.
- Never expose generic SQL write tools to agents.
- All external mutations must follow plan, preview, approval, execute.
- Reverse ETL must require an approval token or equivalent explicit approval artifact.
- All CLI JSON output must be deterministic and versioned.
- Human logs and warnings go to stderr. Machine JSON goes to stdout.
- Any untrusted terminal output must be sanitized for ANSI/control/bidi/zero-width injection.
- Any file path, table name, connector name, URL, stream name, and command argument controlled by users or agents must be validated at the parse boundary.
- Do not add new Go module dependencies unless the phase explicitly requires them and the user approves. Prefer the dependencies already present in go.mod.
- Prefer Podman for runtime services. If Podman is unavailable, fall back to Docker.
- Prefer GHCR images where trusted GHCR images are available.

TDD contract:
For every behavior change, use red-green-refactor.

RED:
- Write the failing unit, integration, golden, CLI, or acceptance test first.
- Run the narrowest command that proves the test fails for the expected reason.
- Record the test name, command, failure summary, and expected behavior in .planning/phases/agentic-etl-platform/TDD-LEDGER.md.

GREEN:
- Implement the smallest production change that passes the failing test.
- Run the same narrow test again.
- Record the passing command in TDD-LEDGER.md.

REFACTOR:
- Improve naming, boundaries, duplication, and Go idioms without changing behavior.
- Re-run the narrow test plus the relevant package tests.
- Record the refactor verification in TDD-LEDGER.md.

Do not batch many behaviors under one vague test. Use focused tests and grow coverage in slices.

Implementation slices:

1. Preflight and baseline
- Run repo profile and inspect the existing architecture.
- Create or update .planning/phases/agentic-etl-platform/PROMPTS.md with this prompt.
- Create or update RUN-STATE.json, TDD-LEDGER.md, PRD-COVERAGE.md, SUMMARY.md, and VERIFICATION.md.
- Run the baseline verification commands before edits and record the result.

2. Structured CLI contracts
- Add typed error categories and stable exit codes for success, usage, auth, validation, connector, runtime, policy, and internal errors.
- Return machine-readable JSON errors when --json is requested.
- Keep human-readable errors concise and sanitized on stderr.
- Add tests for exit codes, JSON error shape, stderr behavior, and unknown command handling.

3. Agent-safe validation and output sanitization
- Add shared validators for connector names, credential names, stream names, table names, paths, URLs, and enum flags.
- Add terminal output sanitization for user/API-sourced strings.
- Add rejection tests for control characters, ANSI escapes, bidi overrides, zero-width characters, path traversal, unsafe absolute paths where disallowed, and invalid identifiers.

4. Connector manifest and registry model
- Introduce a Go connector manifest model that captures metadata, capabilities, config fields, secret fields, streams, primary keys, cursor fields, sync modes, pagination options, and risk classification.
- Adapt existing sample, github, file, warehouse, and outbox connectors to expose manifests.
- Keep connector logic explicit Go code. Do not implement untrusted plugin loading.
- Add tests that manifests drive connector inspection, docs generation, and agent schema output.

5. Generated docs and skills
- Generate detailed man-style docs from the command registry and connector manifests.
- Add pm docs generate support for markdown/json if missing or incomplete.
- Add pm skills generate to create Codex/Claude-compatible SKILL.md files for:
  - pm-shared
  - pm-connectors
  - pm-github
  - pm-etl
  - pm-reverse-etl
  - pm-runtime
  - recipe-github-prs-to-warehouse
  - recipe-preview-approve-reverse-etl
- Add an auto-generated docs/skills index.
- Add AGENTS.md and CONTEXT.md with safe agent usage rules.
- Add tests or golden snapshots for generated docs/skills.

6. Credential hardening
- Preserve dependency-free AES-GCM local vault behavior.
- Add credential source precedence:
  1. explicit env secret fields supplied by command
  2. credential file when explicitly configured
  3. encrypted vault credentials
  4. optional OS keychain-backed vault key with file fallback
- Keep file fallback for CI, containers, and headless environments.
- Validate file permissions where possible and warn safely.
- Add tests proving credential inspection never returns decrypted values.
- Add tests proving logs, JSON output, docs, and generated skills do not contain secret values.

7. Streaming ETL and checkpointing
- Replace full-run buffering with batch streaming.
- Preserve connector emit-style reads, but write transformed records to destinations in bounded batches.
- Add run progress, page progress, record counts, checkpoint state, resumability, and retry-safe metadata.
- Add GitHub pagination controls equivalent to page-all, page-limit, and page-delay while preserving connector-specific config.
- Add tests that large multi-page GitHub-style reads do not buffer all records before destination writes.
- Add tests for resume from checkpoint, failed page handling, cancellation, and context cancellation.

8. Reverse ETL safety model
- Keep reverse ETL plan, preview, approval, and run as separate commands.
- Add policy/risk classification from destination connector manifests.
- Add approval expiration, approval token verification, dry-run support, and per-record receipts.
- Add tests for missing approval, expired approval, invalid token, preview mismatch, and successful approved execution.

9. Runtime-backed mode
- Keep dependency-free mode as the default.
- Use runtime-backed mode only when requested.
- PostgreSQL: durable run ledger, catalog snapshots, approvals, and optional metadata store.
- DragonflyDB: leases, rate limits, ephemeral coordination, and batch pointers.
- Temporal: durable workflow target for ETL/reverse ETL orchestration behind an interface.
- Keep in-process local execution available.
- Runtime scripts must prefer Podman, then Docker.
- Use GHCR images where trusted images are available.
- Add integration tests gated by POLYMETRICS_INTEGRATION=1 so normal go test remains fast.

10. Agent helper workflows
- Add typed helper commands for common safe workflows. Helpers must add orchestration, not merely wrap one raw command.
- Candidate helpers:
  - pm github +sync-prs
  - pm etl +github-prs-to-warehouse
  - pm reverse +preview-approve-run
  - pm runtime +doctor
- Add tests that helpers still respect preview/approval boundaries and redaction rules.

11. Benchmarks and performance comparison
- Maintain dependency-free and runtime-backed performance comparisons.
- Benchmark GitHub pull request ETL across many pages using synthetic HTTP tests and, only when explicitly configured, live API tests.
- Capture elapsed time, records/sec, pages/sec, peak memory where feasible, allocations where feasible, and runtime dependency health.
- Add benchmark docs under docs/performance/.

Required verification commands:
- gofmt on touched Go files
- go vet ./...
- go test ./...
- go build ./cmd/pm
- make verify
- make smoke
- ./pm help
- ./pm docs generate --dir docs/cli
- ./pm skills generate --dir docs/skills or the implemented equivalent

Runtime verification when runtime-backed code is touched:
- scripts/runtime.sh doctor
- scripts/runtime.sh up
- scripts/runtime.sh ps
- POLYMETRICS_INTEGRATION=1 go test ./...
- scripts/runtime.sh down

Performance verification when ETL/runtime execution is touched:
- make perf-free
- If runtime services are available: make runtime-up, make perf-runtime, make runtime-down
- Record results in docs/performance/ and .planning/phases/agentic-etl-platform/VERIFICATION.md.

Acceptance criteria:
- The CLI remains a single Go binary.
- Dependency-free mode works with no PostgreSQL, DragonflyDB, Temporal, Rails, Ruby, or web UI.
- Runtime-backed mode is optional and health-checked.
- Connectors expose manifests and retain explicit Go implementations.
- ETL handles large paginated reads with bounded memory and checkpoint/resume behavior.
- Reverse ETL cannot write externally without preview and approval.
- Credentials can be added via terminal/env without secret leakage.
- Generated docs and skills describe the safe command surface.
- Agent-facing JSON is stable, typed, deterministic, and redacted.
- Tests prove safety invariants, pagination behavior, and approval boundaries.
- All planning artifacts are updated.

Stop and ask for explicit approval before:
- Adding a new Go module dependency.
- Changing credential storage semantics in a way that can make existing credentials unreadable.
- Running live external API tests using real tokens.
- Running destructive reverse ETL against a real external system.
- Reducing or skipping quality gates.
- Deleting user data, credentials, state, or generated benchmark outputs.

Final response required:
- Summarize implemented slices.
- List verification commands and results.
- List any skipped checks with concrete reasons.
- List changed files grouped by area.
- List remaining risks and next recommended phase.
```
