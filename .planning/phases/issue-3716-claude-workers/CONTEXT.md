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
- The base allowlist contains only `Bash`, `Edit`, `Glob`, `Grep`, `Read`, and `Write`. Claude's
  documented `skills` frontmatter preloads the required Go guidance through
  `cc-skills-golang:*` and frontend guidance through `frontend-design:frontend-design`; every
  preload is plugin-qualified. `Skill`, MCP tools, and every other unneeded built-in tool are
  omitted. `disallowedTools` denies `Agent`, its legacy `Task` alias, and `Skill`.
- The official Claude Code subagent documentation states that omitting `Agent` from a tools
  allowlist blocks spawning through that tool and recommends the `skills` field for preloading.
  The official skills documentation states that plugin namespaces cannot collide with personal or
  project skill names and that `context: fork` runs only when a skill is invoked. The policy denies
  runtime `Skill` access, validates every qualified preload, and proves sync restores drift.
- `vercel-composition-patterns`, `vercel-react-best-practices`, and `web-design-guidelines` have no
  trusted plugin-qualified source in the reviewed Claude environment. They remain unavailable to
  this worker; website/docs UI jobs requiring them must preserve state and hand off to a
  captain-approved harness with trusted plugin packaging.
- Claude project definitions are discovered from every `.claude/agents/` scope while walking from
  the current working directory to the repository root, and each scope is scanned recursively. The
  canonical check must inventory all repository-local scopes and reject extra files, duplicate
  names, and symlinks. It prunes only the exact root `.git` metadata directory; `.GIT`, `.Git`,
  and nested directories are ordinary inventory paths. Project definitions outrank same-name user
  and plugin definitions, but managed definitions and CLI `--agents` remain higher-precedence
  caveats.
- The two files must be complete generated projections so every frontmatter or body change fails
  the drift check. The Claude renderer collapses any carriage-return run immediately before LF to
  one LF in a linear pass, preserving bare carriage returns and reaching an idempotent fixed point.
  The check/sync boundary applies the same canonicalization to expected and actual whole-file bytes
  before exact comparison; every other byte remains canonical. The existing
  bounded/symlink-safe projection writer remains the only writer.

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
