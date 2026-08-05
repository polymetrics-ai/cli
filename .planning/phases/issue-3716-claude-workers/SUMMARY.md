---
phase: issue-3716-claude-workers
status: complete
coverage:
  - id: D1
    description: Two generated project-local Claude worker definitions have valid required frontmatter
    verification:
      - kind: unit
        ref: internal/agentcontract/claude_test.go TestClaudeProjectWorkersBlockAmbientAgentDelegation
        status: pass
      - kind: other
        ref: go run ./cmd/agentcontractgen check
        status: pass
    human_judgment: false
  - id: D2
    description: The selected project worker omits direct and runtime skill-mediated delegation routes
    verification:
      - kind: unit
        ref: internal/agentcontract/check_test.go TestProjectionDriftCheckAndSync
        status: pass
      - kind: integration
        ref: Claude Code v2.1.222 forced-Agent smoke with ambient CLI and plugin fixtures
        status: pass
    human_judgment: false
  - id: D3
    description: The canonical source owns full-file generation and fails closed on nested agent or semantic drift with idempotent EOL handling
    verification:
      - kind: unit
        ref: internal/agentcontract/claude_test.go TestRenderClaudeProjectionsIsStableAndSelfContained
        status: pass
      - kind: unit
        ref: internal/agentcontract/check_test.go inventory and canonical/working-tree EOL fixed-point tests
        status: pass
      - kind: other
        ref: make agent-contract-check
        status: pass
    human_judgment: false
  - id: D5
    description: Canonical paths are slash-native and GSD adapter execution is Node-mediated
    verification:
      - kind: unit
        ref: internal/agentcontract/check_test.go TestCheckGSDCommandsRunsFromSelectedRoot
        status: pass
      - kind: unit
        ref: internal/agentcontract/contract_test.go TestCanonicalContractRequiredInvariants
        status: pass
    human_judgment: false
  - id: D4
    description: Documented discovery and precedence caveats are explicit without changing ambient roles or plugins
    verification:
      - kind: other
        ref: .claude/agents/pm-delivery-worker.md generated policy section
        status: pass
      - kind: other
        ref: git diff --name-only origin/refactor/3714-canonical-delivery-flow...HEAD
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3716 clean project-local Claude workers

## Delivered

- Added a validated Claude harness policy to the one canonical delivery contract and made only the
  two Wave 2 Claude projections required.
- Added deterministic full-file Markdown/YAML rendering and whole-file drift checking/sync for
  `.claude/agents/pm-delivery-worker.md` and `.claude/agents/pm-connector-worker.md`.
- Generated both workers from the canonical source. Their explicit allowlist contains only the six
  base tools; `skills` preloads trusted plugin-qualified Go/frontend guidance. The runtime `Skill`
  tool is omitted, while `disallowedTools` denies `Agent`, `Task`, and `Skill`.
- Recorded `vercel-composition-patterns`, `vercel-react-best-practices`, and
  `web-design-guidelines` as unavailable because they lack a trusted plugin-qualified source. A
  website/docs UI job requiring them must hand off to an approved harness rather than weakening
  isolation.
- Added red/green regression coverage for missing project workers, invalid `Agent` drift, atomic
  sync repair, frontmatter requirements, repository-wide nested extra/duplicate agent inventory
  including case-variant Git-like paths, canonical and working-tree EOL fixed-point stability with
  repeated carriage returns, and connector inheritance.
- Made slash-separated contract path validation independent of the host OS and execute the
  extensionless GSD JavaScript adapter through Node.
- Recorded the official Claude Code discovery, precedence, and tool-access rules together with an
  honest live smoke result and the managed/CLI override limitation.

## TDD outcome

RED failed before generation because no project-local workers existed, leaving an ambient
same-name worker eligible to be selected. GREEN generated required project files with the explicit
no-`Agent` policy. Review hardening replaced collision-prone skill rules with qualified preloads,
denied runtime `Skill`, and added an `Agent`/`Task`/`Skill` denylist so unrelated
`context: fork` skills cannot be invoked. Repository-wide inventory and EOL-normalized drift tests
demonstrate that policy is canonical and repaired by the actual generator. Only exact root `.git`
metadata is pruned, while the renderer and validator share idempotent LF-canonical whole-file
semantics even for repeated carriage-return runs. The prior trusted-home Claude Code v2.1.222
smoke forced an `Agent` call against ambient fixtures and returned `AGENT_UNAVAILABLE` without
tool use.

## Isolation boundary

The official sources are https://code.claude.com/docs/en/sub-agents and
https://code.claude.com/docs/en/slash-commands. They document project agent precedence, `Agent`
omission, the subagent `skills` preload field, plugin namespace collision resistance, and
`context: fork`. The generated workers preload only qualified plugin identifiers and deny `Agent`,
`Task`, and `Skill`. The prior live smoke proves direct `Agent` is absent from the selected worker;
it does not prove clean-home selection, runtime preload resolution, plugin source/version pinning,
or prevent a higher-precedence managed/CLI same-name agent definition.

## Scope outcome

No installed plugin, global home configuration, `.codex`, `.pi`, legacy worker, connector bundle,
or connector overlay content changed. The clean-home runtime smoke is **NOT PERFORMED**; the
unauthenticated fixture is not runtime evidence, and no credentials were accessed or copied.
