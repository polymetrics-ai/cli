# Subagent Extension

This repository keeps a small fork of Pi's [official subagent extension example](https://github.com/earendil-works/pi/tree/main/packages/coding-agent/examples/extensions/subagent). Pi roles are supplied by that extension model; this fork does not load a bundled agent roster.

It registers the `subagent` tool and keeps `/sub-agent-settings` for explicitly selected user roles. The generated clean-project workers never use that settings command.

## Clean project workers

`agentScope: "clean-project"` is the default. It reads the repository's canonical delivery contract and loads exactly these generated project files:

- `.pi/agents/pm-delivery-worker.md`
- `.pi/agents/pm-connector-worker.md`

The clean scope runs `agentcontractgen check` before loading workers and fails closed when the canonical contract or either expected generated file is missing, malformed, drifted, or cannot pass that validation through the repository's Go toolchain. It does not scan user roles, the extension's historical `agents/` directory, or the retained legacy `.pi/agents` roles. Those legacy files remain until their separately gated migration wave.

The extension still supports the official-example scopes when they are explicitly selected:

| `agentScope` | Discovery behavior |
| --- | --- |
| `clean-project` (default) | Only the two canonical generated project workers above. |
| `user` | `~/.pi/agent/agents/*.md`. |
| `project` | Every role in the nearest `.pi/agents/*.md`. |
| `both` | User roles followed by project roles; matching project names override user names. |

The user/project locations, YAML frontmatter parser, and user-then-project precedence match the [official discovery example](https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/examples/extensions/subagent/agents.ts). Pi's [settings documentation](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/settings.md) covers project trust and the non-interactive trust behavior. This extension keeps its confirmation requirement for every project-aware scope, including `clean-project`.

## Generated Pi agent format

The two project files are complete generated projections of `.agents/agentic-delivery/canonical/delivery-contract.json`; never edit them directly. They use the Markdown-with-YAML shape used by Pi's official example:

```markdown
---
name: "pm-delivery-worker"
description: "..."
tools:
  - "read"
  # more bounded tools
# optional: model: "provider/model"
---

Generated system prompt
```

`name` and `description` are required. `tools` and `model` are optional Pi frontmatter fields; this repository deliberately generates `tools` for both workers from the canonical bounded allowlist. See the [official example parser](https://raw.githubusercontent.com/earendil-works/pi/main/packages/coding-agent/examples/extensions/subagent/agents.ts).

## Delegation boundary

The `subagent` tool supports single, parallel, and chain work. It retains the fork's bounded child-tool intersection and depth check: `subagent` is removed from every child allowlist and `PI_SUB_AGENT_DEPTH=1` causes a nested request to fail before another process is spawned.

Each child invocation also adds Pi's documented `--no-extensions` option, in addition to `--tools` and `--no-session` (or the recordable session-dir form). Pi documents `--tools` as an allowlist and `--no-extensions` as disabling extension loading in its [coding-agent README](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/README.md). Therefore a clean worker cannot reach this extension's `subagent` tool or an ambient plugin-provided delegation tool in its child process.

This is delegation isolation, not a general execution sandbox: the generated workers intentionally retain their bounded built-in tools, including `bash`, so they can perform the canonical delivery flow. Project trust and the confirmation gate still apply before a project worker is started.

## Parameters

| Field | Applies to | Notes |
| --- | --- | --- |
| `agent` + `task` | Single | Run one named agent for one task. |
| `tasks` | Parallel | Array of `{ agent, task, cwd? }`; maximum 8 tasks and 4 concurrent subprocesses. |
| `chain` | Chain | Array of `{ agent, task, cwd? }`, maximum 8; `{previous}` is replaced with prior output. |
| `agentScope` | All | Defaults to `clean-project`; the other scopes are explicit compatibility modes. |
| `confirmProjectAgents` | All | Defaults to `true`; without UI, project workers are blocked unless a trusted caller explicitly sets `false`. |
| `cwd` | All | Default working-directory override; task and chain entries can override it. |

Delegated task text is sent over stdin, not process arguments. Parallel work is capped at 8 tasks with 4 child processes; chains are capped at 8 steps and stop at their first failed step. Child output is bounded for parent context while structured details retain the full result for rendering.

## Verification

Run `bash scripts/tests/pi-clean-project-agents.sh` where Pi and Bun are installed. It imports the real extension and proves that clean discovery exposes only the two generated project workers despite a fixture global role, the extension's historical roster, and retained legacy project roles. It also checks rejection of parseable but drifted contract and worker files, the generated bounded tool list, the removed `subagent` tool, the depth block, and the `--no-extensions` child argument.

Run `go run ./cmd/agentcontractgen sync` to create or refresh the generated Pi workers, then `go run ./cmd/agentcontractgen check` to enforce exact full-file drift detection.
