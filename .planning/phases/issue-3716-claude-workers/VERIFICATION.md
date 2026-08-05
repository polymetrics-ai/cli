# VERIFICATION — issue #3716 clean project-local Claude workers

Status: implementation verification in progress; delivery gates remain.

## Checklist

- [x] Exactly two required `.claude/agents` projections exist.
- [x] Both are complete generated files derived from the canonical source.
- [x] YAML frontmatter has required `name` and `description`, explicit minimum tools, and
  `permissionMode: default`.
- [x] Neither worker grants `Agent`, `Skill`, MCP, an orchestration persona, or an ambient-agent
  escape path.
- [x] Mutating a generated file to grant `Agent` fails the real drift check; sync restores it.
- [x] Installed Claude CLI smoke discovers both project workers by name in this trusted checkout.
- [x] Live ambient CLI/plugin fixtures demonstrate that delegation is blocked for the selected
  project worker; a clean-home same-name user fixture was intentionally not authenticated, so
  user-home precedence is documented/static rather than model-executed.
- [x] Documentation records project discovery, precedence, tool allowlisting, and the managed/CLI
  override caveat using https://code.claude.com/docs/en/sub-agents.
- [x] No global Claude, installed plugin, `.codex`, `.pi`, legacy-role, or connector path changed.
- [ ] Focused tests, canonical check, local gates, no-mistakes child gates, and sub-PR CI are green.

## Isolation evidence and boundary

The generated YAML `tools` field is an explicit allowlist and deliberately does not contain
`Agent` (or `Skill`/MCP tool names). The official Claude Code subagents documentation states that
omitting `Agent` means the subagent cannot spawn subagents. The focused regression began RED when
the project definitions were absent, then verifies the generated no-`Agent` files and drift repair.

Claude Code v2.1.222 was also run from this trusted repository with an ambient CLI fixture and a
temporary local plugin fixture available. A prompt requiring `pm-delivery-worker` to use `Agent`
to invoke either fixture returned `AGENT_UNAVAILABLE` without a tool invocation. This is direct
runtime evidence that the selected project worker cannot delegate to ambient agents.

The same-name user fixture was created only under a temporary clean `HOME`. Claude Code correctly
had no login there and stopped before model execution. No real-home credentials were read, copied,
or supplied. Therefore this phase does **not** claim a clean-home runtime precedence execution;
the user/plugin precedence ordering is the documented behavior, and the project file's full-file
drift check is static evidence. Managed definitions and CLI `--agents` are documented as higher
precedence than project roles, so they can intentionally replace a same-name project definition;
this harness cannot prevent that override.
