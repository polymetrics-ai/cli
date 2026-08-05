---
phase: issue-3715-pi-project-workers
status: complete
coverage:
  - id: D1
    description: Canonical source generates exactly the two complete Pi workers and rejects whole-file drift.
    verification:
      - kind: unit
        ref: internal/agentcontract/check_test.go TestSyncCreatesRequiredPiProjections
        status: pass
      - kind: unit
        ref: internal/agentcontract/check_test.go TestPiProjectionRejectsWholeFileDrift
        status: pass
      - kind: other
        ref: go run ./cmd/agentcontractgen check
        status: pass
    human_judgment: false
  - id: D2
    description: Default clean-project discovery exposes only the canonical delivery and connector workers.
    verification:
      - kind: integration
        ref: bash scripts/tests/pi-clean-project-agents.sh
        status: pass
    human_judgment: false
  - id: D3
    description: A clean worker cannot delegate through ambient extensions or recursive subagents.
    verification:
      - kind: integration
        ref: bash scripts/tests/pi-clean-project-agents.sh
        status: pass
    human_judgment: false
  - id: D4
    description: Project confirmation and the canonical bounded tool policy remain fail-closed.
    verification:
      - kind: integration
        ref: bash scripts/tests/pi-clean-project-agents.sh
        status: pass
      - kind: unit
        ref: internal/agentcontract/contract_test.go TestCanonicalContractRequiredInvariants
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3715 Pi clean project-only workers

## Delivered

- Extended the Wave 1 canonical contract with an exact two-role Pi projection and its bounded
  built-in child-tool declaration.
- Generated `.pi/agents/pm-delivery-worker.md` and `pm-connector-worker.md` as complete
  Markdown/YAML files; the drift check now creates them when absent and rejects any whole-file
  hand edit.
- Removed the extension-bundled roster from discovery. The new default `clean-project` mode reads
  the canonical contract, accepts only the two exact regular project files, and ignores global,
  historical bundled, and retained legacy project roles.
- Preserved the recursive-depth guard and bounded child-tool intersection, then added
  `--no-extensions` to every child Pi invocation so no extension can register an ambient
  delegation tool in a worker process.
- Documented Pi's official frontmatter/discovery/tool behavior, project trust, scope precedence,
  and the precise delegation-isolation boundary.

## TDD outcome

RED demonstrated the original loader exposed the hostile global fixture, all nine bundled roles,
and retained legacy project roles; a second RED test showed required Pi projections were optional
and could not be created by `sync`. GREEN makes both tests pass with generated full files and a
fail-closed clean scope. REFACTOR extracted child policy, rejected worker symlinks and unsafe role
file stems, and added whole-file and `os.Root` ancestor-symlink checks.

## Scope outcome

No legacy `.pi/agents` role was deleted. No global Pi/Codex/Claude state, `~/.pi/agent/shepherd`,
GSD adapter, GSD skill, connector behavior, or dependency changed.
