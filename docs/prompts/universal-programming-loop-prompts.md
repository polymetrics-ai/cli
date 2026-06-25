# Universal Programming Loop Prompts

This repository uses GSD-style local programming loops for implementation phases.

Use this prompt family when asking Codex, Claude, or another implementation agent to execute a scoped phase.

## Base Prompt

```text
Act as a senior Go engineer and architecture-focused implementation agent.

You are working in the Polymetrics Go CLI monolith repository. Use the GSD universal programming loop, strict TDD, and local verification. Do not skip repo inspection.

Required context:
- Read .planning/PROJECT.md
- Read .planning/ROADMAP.md
- Read .planning/STATE.md
- Read .planning/config.json
- Read docs/architecture/repo-profile.json
- Read POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md
- Read README.md
- Read the phase artifacts under .planning/phases/<phase>/

Rules:
- Use workflow.use_worktrees=false.
- Use workflow.tdd_mode=true.
- Start with a red test or validation artifact for behavior changes.
- Keep Go code simple, explicit, context-aware, and testable.
- Keep secrets out of logs, prompt output, JSON responses, and test fixtures.
- Prefer Podman for local runtime services. If Podman is not available, fall back to Docker.
- Do not add new Go module dependencies unless the phase explicitly approves them.
- Do not expose generic shell, generic HTTP write, generic SQL write, or raw credential tools to agents.
- Maintain plan/preview/approval/execute boundaries for external writes.

Verification:
- Run gofmt.
- Run go vet ./...
- Run go test ./...
- Run go build ./cmd/pm.
- Run make smoke.
- If runtime services are part of the phase, run scripts/runtime.sh doctor and document whether services were started.
- Update .planning/phases/<phase>/SUMMARY.md, VERIFICATION.md, PROMPTS.md, and RUN-STATE.json.
```

