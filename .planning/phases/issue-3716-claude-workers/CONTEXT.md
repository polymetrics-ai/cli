# CONTEXT — issue #3716 clean project-local Claude workers

## Phase mapping

GitHub sub-issue #3716 (Wave 2 of parent #3714) maps to this GSD phase. It depends on Wave 1
(#3718), whose canonical source and projection generator are already on the parent branch. The
captain accepted Wave 1 PR #3724 after 23 successful checks, 6 skipped checks, and no failures.
That acceptance explicitly records no separate automated code-review pass and keeps the resulting
coverage gap visible without retrofitting coverage or blocking Wave 2.

## Locked decisions

- Create exactly two generated project-local Claude Code definitions:
  `.claude/agents/pm-delivery-worker.md` and `.claude/agents/pm-connector-worker.md`.
- The canonical JSON source remains the only delivery-flow source. Claude-native frontmatter policy
  is modeled there and rendered by `cmd/agentcontractgen`; generated files are never hand-edited.
- Claude Code project definitions use Markdown with YAML frontmatter. `name` and `description` are
  required; an explicit `tools` allowlist and `permissionMode: default` are rendered.
- The base allowlist contains `Bash`, `Edit`, `Glob`, `Grep`, `Read`, and `Write`, plus scoped
  `Skill(name)` entries for the Go and design skills required by
  `required-skills-routing.md`. Bare `Skill`, MCP tools, and every other unneeded built-in tool are
  omitted. `disallowedTools` denies `Agent` and its legacy `Task` alias.
- The official Claude Code subagent documentation states that omitting `Agent` from a tools
  allowlist blocks spawning through that tool. The official skills documentation also states that
  bare `Skill` reaches project, user, plugin, and bundled skills, and that `context: fork` runs a
  skill in a subagent. The policy therefore grants only named skill rules and excludes unrelated
  fork-capable skills. The test must enforce this policy and prove sync restores drift.
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
