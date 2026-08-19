# CONTEXT — issue #3715 Pi clean project-only workers

## Phase mapping

GitHub sub-issue #3715 (Wave 4 of parent #3714) maps to this GSD phase. It depends on the
Wave 1 canonical delivery contract merged in #3724 and delivers its two Pi projections.

## Locked decisions

- The only active clean-project roles are generated `pm-delivery-worker` and its inheriting
  `pm-connector-worker` overlay. Their Markdown/YAML frontmatter and body are rendered from
  `.agents/agentic-delivery/canonical/delivery-contract.json`; generated files are never edited by
  hand.
- Pi core has no built-in subagent facility. This repository's `subagent` capability is a
  project-local fork of Pi's official extension example. Official references:
  - https://github.com/earendil-works/pi/blob/main/packages/coding-agent/README.md
  - https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/examples/extensions/subagent/agents.ts
  - https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/examples/extensions/subagent/index.ts
  - https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/docs/extensions.md
- The extension gets an explicit `clean-project` discovery scope. It reads the two allowed role
  names from the canonical contract, loads no user-global or extension-bundled agents, and ignores
  retained legacy project roles. The extension's default scope is `clean-project`.
- The bundled agent roster is no longer loaded by any scope. Existing `.pi/agents` legacy files
  stay intact for Wave 6; clean-project mode simply cannot discover them.
- The canonical Pi worker frontmatter uses a bounded tool allowlist that excludes `subagent`. The
  extension preserves its runtime defense-in-depth removal of `subagent` from every child
  allowlist and its `PI_SUB_AGENT_DEPTH` recursive-delegation block.
- Project trust and confirmation safeguards remain: clean-project is a project scope and retains
  the existing confirmation behavior, including the non-interactive refusal unless explicit trust
  is supplied.

## Scope fences

- Do not delete the legacy `.pi/agents` role fleet; that is Wave 6.
- Do not touch `~/.pi/agent/shepherd` or any global Pi, Codex, or Claude configuration/state.
- Do not alter `.gsd`, `.pi/extensions/gsd`, or `.pi/skills/gsd-core`.
- Do not add dependencies or modify connector/runtime behavior.

## GSD execution note

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop ...` returned `unknown GSD command`; Wave 1 established
  that the installed adapter intentionally has no such command. This phase follows the
  documented inline/manual GSD fallback without relaxing TDD, verification, or review.
- Executed generated prompts: `discuss-phase issue-3715-pi-project-workers --auto` and
  `plan-phase issue-3715-pi-project-workers --tdd --skip-research`.
