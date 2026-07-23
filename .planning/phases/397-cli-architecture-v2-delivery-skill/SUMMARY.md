# CLI Architecture v2 delivery skill summary

## Outcome

Added the approved narrow `cli-architecture-v2-delivery` program router on top of
`origin/feat/cli-architecture-v2`. It separates parent implementation truth from default-branch
delivery, defines stable scheduling states, enforces current remote/exact-head evidence, routes each
track to existing specialist skills, and keeps parent promotion and merge human/orchestrator-owned.

The existing `bubble-tea-tui-design` skill remains the sole detailed TUI authority. This slice added
only its approved cross-link. No generic Go TUI skill, product command/UI behavior, runtime
dependency, help/manual/website output, shared parent orchestration state, or GitHub issue state was
changed.

## Files and routing

- `.agents/skills/cli-architecture-v2-delivery/SKILL.md`
- three stable references for state/dependencies, phase delivery, and parent review/integration
- `agents/openai.yaml` discovery metadata
- `AGENTS.md`, required-skills routing, and task-skill matrix force triggers
- focused `scripts/tests/cli-architecture-v2-delivery-skill.sh`
- focused `make cli-architecture-v2-skill-check`, intentionally excluded from global `make verify`
- dated research/source and issue-gap evidence in this phase directory

## GSD/TDD evidence

The repo-local adapter was healthy with 69 commands. `programming-loop` remained absent, so the
manual universal loop was recorded and followed:

1. corrected plan and audit/adoption decision;
2. RED focused contract at commit `8a459be64` (missing skill, exit 1);
3. GREEN skill/routing at `d81163f70`;
4. independent findings fixed at `2ba47dc30`;
5. focused, broad, independent exact-head, and Shepherd verification.

Independent review found two initial defects: omitted conditional benchmark/performance routing for
Track C and a Darwin-concurrent `mktemp` collision. Both were fixed; eight concurrent checker
invocations subsequently passed. Fresh review of `2ba47dc30` reported no actionable findings. A
bounded Shepherd-style trajectory review scored 4.52, returned `PROCEED`, and is persisted in
`SHEPHERD-VALIDATION-2BA47DC.json`, explicitly authorizing the planning-artifact refresh.

A later final-head review found incomplete negative validation, missing universal-loop decision
fields, and an overstated exact-head Shepherd claim. This fix round parses skill frontmatter as YAML,
checks the complete Make `verify` rule including recipes, records each orchestration decision and
verification state, and limits the persisted Shepherd claim to the head actually reviewed. Final
Codex and trajectory reruns are required on the immutable fix head and will be recorded in the draft
stacked PR. A subsequent re-review found that duplicate Make `verify` declarations and the required
universal-loop cycle vocabulary still needed coverage; the final fix round inspects all duplicate
rules and records the six required cycles with allowed decisions.

## Verification

Passed:

- `scripts/tests/cli-architecture-v2-delivery-skill.sh`
- concurrent direct and Make focused checks
- `make cli-architecture-v2-skill-check`
- `scripts/tests/pi-model-routing.sh`
- `bash -n` and independent `shellcheck` of the focused script
- independent YAML/frontmatter and local-link validation
- `git diff --check`
- `go vet ./...`
- `go test -timeout 20m ./...`
- `go build ./cmd/pm`
- `make verify`, including docs, smoke, lint, connector validation, and Shepherd guard self-tests

`go.mod`, `go.sum`, Go/product code, generated/user help surfaces, design/ADR docs, website content,
and shared parent #397 orchestration artifacts remained unchanged from the parent base.

## Delivery state

The worker branch remains based on `feat/cli-architecture-v2`. Because the installed no-mistakes
version hardcodes `origin/main`, the captain approved direct stacked delivery after equivalent green
checks, fresh Codex review, and exact-head trajectory validation. The draft PR must use `Refs #397`,
exclude configured review bots, and target only the parent. Parent PR #438 remains draft and must not
be merged by an agent.
