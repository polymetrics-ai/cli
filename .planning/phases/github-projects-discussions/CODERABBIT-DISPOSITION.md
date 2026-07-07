# CodeRabbit PR #74 Review Disposition

Review source: `coderabbitai` on PR #74 (run ID `4e53a71c-2046-449b-aa01-cc0e3bb97d55`).

## Disposition legend

- **accepted**: implemented in this branch.
- **declined**: still-valid finding but out of scope, already fixed, or intentionally not changed; reason recorded.
- **deferred**: valid but requires run evidence or human decision; follow-up issue recommended.
- **duplicate**: same issue covered by another comment.

---

## Actionable comments

### Agentic-delivery / docs

| # | File | Finding | Classification | Reason / Change |
|---|------|---------|----------------|-----------------|
| 1 | `.agents/agentic-delivery/references/caveman-token-compression.md:47-48` | Hard-coded personal home path in installed-skill example. | **accepted** | Replaced `/Users/karthiksivadas/...` with `~/.agents/...`. |
| 2 | `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md:32-34` | Fallback decision token `failed_runtime_capability` outside required vocabulary. | **accepted** | Replaced with `not_spawned_runtime_capability_missing` to match the spawn-decision contract. |
| 3 | `.agents/connector-migration/agents/implementation/passb-expander.agent.yaml:56-73` | Contradictory commit policy: step allows green-slice commits, guardrail forbids any commits. | **accepted** | Aligned guardrail to "commit green slices only after local gates pass and the connector is assigned exclusively to this agent". |
| 4 | `docs/prompts/universal-programming-loop-prompts.md:38-45` | Spawn-decision labels use hyphenated prose instead of canonical tokens. | **accepted** | Updated to `spawned`, `read_only_spawned`, `local_critical_path`, or one `not_spawned_*` blocker. |

### GitHub connector / engine

| # | File | Finding | Classification | Reason / Change |
|---|------|---------|----------------|-----------------|
| 5 | `internal/connectors/defs/github/operations.json:33-42` | `ListProjectItems` missing `DraftIssue` fragment in content union. | **accepted** | Added `... on DraftIssue { id title }` to the content union so draft project items return fields. |
| 6 | `internal/connectors/engine/bundle.go:1023-1036` | GraphQL variable `default` not validated against declared `type`. | **accepted** | Added `validateGraphQLDefaultForType` rejecting mismatched `integer`, `number`, and `boolean` defaults. |
| 7 | `internal/connectors/engine/read_test.go:230-266` | No test for explicitly-empty (not absent) query variable. | **accepted** | Added `TestReadGraphQLBodyOmitsExplicitlyEmptyQueryVariable` asserting omission when `omit_when_empty=true`. |

### Generated / metadata noise

| # | File | Finding | Classification | Reason / Change |
|---|------|---------|----------------|-----------------|
| 8 | `website/next-env.d.ts:1-6` | Auto-generated Next.js file toggles between dev/build paths; should not be committed. | **accepted** | Added `next-env.d.ts` to `website/.gitignore` and removed the tracked file. |
| 9 | `website/data/connectors.generated.json:36831-36930` | Project/Discussion auth notes are PAT-only and can mislead GitHub App users. | **accepted** | Updated `project list`, `project item-list`, `discussion list`, and `discussion view` notes to distinguish PAT scopes from GitHub App installation permissions. |
| 10 | `docs/architecture/repo-profile.json:76-182` | `.next` build output scanned as separate applications, creating noisy duplicates. | **accepted** | Removed all application entries whose root contains `.next`, keeping only the source-root `cli-polymetrics-ai` entry at `website`. |

### Phase planning artifacts

| # | File | Finding | Classification | Reason / Change |
|---|------|---------|----------------|-----------------|
| 11 | `.planning/phases/github-projects-discussions/agents/*.md` | Scope / Allowed Tools / Inputs / Outputs still `TBD`. | **accepted** | Filled all eight agent role contracts with concrete responsibilities, tools, inputs, and outputs. |
| 12 | `.planning/phases/github-projects-discussions/traces/*.md` | Trace sections still `TBD`, making artifacts non-auditable. | **accepted** | Populated traces with references to actual phase artifacts (PLAN, TDD-LEDGER, VERIFICATION, RUN-STATE, THREAT-MODEL) and marked them DRAFT pending final cross-check against run logs. |
| 13 | `.planning/phases/github-projects-discussions/TEST-PLAN.md:3-13` | Missing conformance gate for `read_query` replay path. | **accepted** | Added explicit `read_query` fixture replay coverage gate and default/type-mismatch gate. |

---

## Additional issues found during review-fix (not from CodeRabbit)

| File | Finding | Classification | Reason / Change |
|------|---------|----------------|-----------------|
| `.pi/prompts/*.md` | Prompt templates use `{{task}}`/`{{target}}` placeholders, which pi does not substitute. | **accepted** | Replaced with pi-supported `$@`/`$1` syntax and added `argument-hint` frontmatter. |
| `.pi/README.md` | Launch command does not enable `grep`/`find`/`ls`, so read-only subagents lose search tools. | **accepted** | Documented `pi --tools read,bash,edit,write,grep,find,ls,subagent --approve`. |
| `.pi/agents/pm-coderabbit-disposition.md` | Agent requests `bash` but contract says it must not post/resolve/push changes. | **accepted** | Removed `bash`; agent is now read-only and the orchestrator provides review records in the task. |
| `.pi/prompts/pm-orchestrate.md` | No `agentScope`, `confirmProjectAgents`, or per-worker `cwd` guidance. | **accepted** | Added Pi runtime constraints section covering scope, confirmation, isolation, concurrency caps, and tool allowlist. |
| `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md` | Missing Pi-specific orchestration adapter. | **accepted** | Created adapter mirroring `codex-active-orchestration-loop.md` with concrete Pi limits and rules. |

---

## Deferred items

None. All CodeRabbit findings and pi-runtime audit findings were addressed in this slice.

## Verification

- `go test ./internal/connectors/engine ./cmd/connectorgen -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — no GitHub findings.
- Full `make verify` — pending final run after all slices (recorded in VERIFICATION.md).
