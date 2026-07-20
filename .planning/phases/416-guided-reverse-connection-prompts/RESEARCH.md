# Phase 18 progressive terminal setup research

Date: 2026-07-20
Issues: #416 (reverse), #469 (credentials/connections)
Design gate: #462 / PR #468

## Question

How should Polymetrics make credential and connection setup intuitive and interactive by default
without weakening its agent-first, pipe-safe, secret-safe CLI contract?

## Primary evidence

| Source | Relevant pattern | Polymetrics decision |
|---|---|---|
| [Bubble Tea v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md) | Declarative `tea.View`; deterministic Model/Update/View loop | Keep state and rendering pure; perform service I/O and cancellation-aware work in `tea.Cmd` |
| [Huh](https://github.com/charmbracelet/huh) | Composable groups, dynamic forms, Bubble Tea embedding, first-class accessible mode | Use for bounded missing-field forms; accessible mode becomes sequential prompts only after the same dual-TTY gate |
| [Bubbles](https://github.com/charmbracelet/bubbles) | Reusable list, help, text input, viewport, and key components | Reuse approved primitives and semantic styles instead of bespoke controls |
| [GitHub CLI `repo create`](https://cli.github.com/manual/gh_repo_create) | Missing arguments can start an interactive flow; flags preserve automation | Apply it to incomplete setup actions and the two explicitly allowlisted place-like bare workspaces (`query`, `reverse`), never as a blanket namespace rule |
| [GitHub CLI environment](https://cli.github.com/manual/gh_help_environment) | Explicit prompt disable and accessible prompter controls | Preserve `--no-input`, CI, PM_NO_TUI, and accessible-mode escape paths |
| [GitHub CLI accessibility guide](https://accessibility.github.com/documentation/guide/cli/) | Numbered static prompts, accessible colors, textual progress | Every interactive choice has words/numbers and a sequential equivalent; color is never the only signal |
| [CLI Guidelines](https://clig.dev/) | Prompt only with interactive stdin; always provide complete flag alternatives; support `--no-input` and structured output | Require both stdin and stdout TTYs, keep flags authoritative, and document `--json --no-input` for agents |
| [Pulumi non-interactive command](https://www.pulumi.com/docs/iac/cli/commands/pulumi_do/) | Explicit non-interactive operation and structured automation output | Retain deterministic prompt-free invocation rather than inventing an overlapping global agent mode |

## Decision matrix

| Invocation | Eligible terminal | Result |
|---|---:|---|
| Bare `pm credentials` or `pm connections` | any | Contextual help, exit 0; never a wizard |
| Bare `pm reverse` | stdin TTY + stdout TTY + no bypass | Open the guided reverse workspace |
| Bare `pm reverse` | pipe, CI, dumb terminal, `--json`, `--plain`, `--no-input`, or PM_NO_TUI | Contextual help, exit 0; never initialize a wizard |
| `pm reverse guide` | stdin TTY + stdout TTY + no bypass | Open the same guided reverse model as bare `pm reverse` |
| Complete action command | any | Execute directly through the existing flag path |
| Incomplete action command | stdin TTY + stdout TTY + no bypass | Launch missing-field guidance |
| Incomplete action command | pipe, CI, dumb terminal, `--json`, `--plain`, `--no-input`, or PM_NO_TUI | Actionable deterministic missing-flag error |
| Complete but invalid action command | any | Field-specific validation error; do not open a repair wizard |

This is progressive enhancement, not a second API. It keeps shell scripts, agents, CI, and tests
deterministic while making a human's incomplete action discoverable.

## Security and recovery findings

- Credential guidance may collect name, connector, auth mode, non-secret connector config, and
  secret-source metadata. It must never accept plaintext secret bytes into a TUI model.
- Existing `--from-env field=ENV` remains the preferred environment handoff. Controlled stdin uses
  a sanitized `--value-stdin field` command shown after the wizard; it is not consumed or saved in
  the Bubble Tea session.
- Connector schema drives required non-secret fields. Literal documentation placeholders such as
  `YOUR_GITHUB_OWNER` must fail local validation before network activity.
- Connection choices come from current credentials, connector definitions, catalogs, capabilities,
  streams, and sync metadata. The review step exposes stream, mode, cursor, key, and destination.
- Duplicate names are ordinary user-correctable validation failures, not internal errors. Recovery
  is inspect, choose another name, or cancel. Never infer overwrite, replace, update, or delete.
- Reverse remains a distinct security-critical session under #416 and preserves plan, preview,
  approval, typed confirmation, execute, and approval-token nondisclosure.

## Agent invocation contract

Documentation should teach agents and automation to use:

```text
--json --no-input
```

Long-running commands may add `--progress ndjson`, whose progress belongs on stderr. Do not add a
global `--agent-mode`: `pm query run --agent-mode summary|stream` already controls query result
shape and must not silently acquire prompt, secret, or output-mode semantics.

## Issue split

The combined Phase 18 scope was too broad for the one-issue/one-PR delivery contract. #416 now owns
the human-first `pm reverse` workspace and its `pm reverse guide` alias; child #469 owns credential
and connection setup. Both remain blocked by the
reviewed #462 design gate, and #469 is also blocked by #409. #417 and #418 explicitly wait for #469.
