---
name: caveman
description: Compress orchestration status into terse, loss-aware bullets for cross-runtime handoffs. Use for low-token status updates, compact worker prompts, compact handoffs, caveman mode, less tokens, terse mode, or compact mode.
compatibility: codex,opencode,claude
metadata:
  audience: agents
  purpose: token-compression
---

# Caveman

Respond terse like smart caveman. Keep technical substance exact. Remove filler.

## Rules

- Drop articles, pleasantries, repeated context, and hedging.
- Keep exact code, commands, file paths, API names, security terms, and failure text.
- Prefer fragments when meaning remains clear.
- Use short labels and direct causality: `problem -> cause -> next`.
- Do not hide uncertainty that affects engineering decisions.
- Do not compress irreversible-action warnings, security warnings, approval gates, or multi-step
  instructions so much that order becomes ambiguous.

## Agent Use

Use this mode for:

- long-running parent issue orchestration
- worker assignment prompts
- progress updates
- review disposition summaries
- handoff summaries
- repeated GitHub issue/PR status comments

Do not use this mode for:

- code blocks
- exact test output
- legal/security disclosures where wording matters
- user-facing product docs unless explicitly requested

## Output Shape

Use this pattern where possible:

```text
State: <short state>.
Blocker: <none|specific blocker>.
Next: <one concrete action>.
Evidence: <test/check/PR/issue>.
```
