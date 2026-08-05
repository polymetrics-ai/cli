# CONTEXT — issue #3716 clean project-local Claude workers

## Phase mapping

GitHub sub-issue #3716 (Wave 2 of parent #3714) maps to this GSD phase. It depends on Wave 1
(#3718), whose canonical source and projection generator are already on the parent branch.

## Locked decisions

- Create exactly two generated project-local Claude Code definitions:
  `.claude/agents/pm-delivery-worker.md` and `.claude/agents/pm-connector-worker.md`.
- The canonical JSON source remains the only delivery-flow source. Claude-native frontmatter policy
  is modeled there and rendered by `cmd/agentcontractgen`; generated files are never hand-edited.
- Claude Code project definitions use Markdown with YAML frontmatter. `name` and `description` are
  required; an explicit `tools` allowlist and `permissionMode: default` are rendered.
- The allowlist contains only `Bash`, `Edit`, `Glob`, `Grep`, `Read`, and `Write`. It intentionally
  omits `Agent`, `Skill`, MCP tools, and every other unneeded built-in tool.
- The official Claude Code subagent documentation states that omitting `Agent` from a tools
  allowlist blocks all subagent spawning. The test must enforce this by mutating a generated
  worker to add `Agent`, demonstrating the ambient-agent reachability regression before sync
  restores the canonical file.
- Claude project definitions are discovered from `.claude/agents/` while walking upward from the
  current working directory. They outrank same-name user and plugin definitions, but managed
  definitions and CLI `--agents` remain higher-precedence caveats.
- The two files must be exact full-file projections so frontmatter changes also fail the drift
  check. The existing bounded/symlink-safe projection writer remains the only writer.

## Scope fences

- Do not modify global `~/.claude`, any installed plugin, plugin cache, `.codex`, `.pi`, legacy
  role files, or connector code.
- Do not create an orchestrator, shepherd, planner, reviewer, verifier, GSD role, or extra worker.
- Do not add a dependency. `gopkg.in/yaml.v3` is already a direct module dependency and may be
  used only if needed to validate generated YAML frontmatter.
- Do not mark parent PR #3723 ready or merge any PR.

## GSD execution note

The repo-local adapter resolves `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`,
`code-review`, and `ship`; it deliberately does not expose `programming-loop`. The generated
`discuss-phase` and `plan-phase --tdd` prompts are executed inline because the canonical contract
forbids role spawning. This preserves the installed lifecycle and TDD evidence without inventing
the absent command.
