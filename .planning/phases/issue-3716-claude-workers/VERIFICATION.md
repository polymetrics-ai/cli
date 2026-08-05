# VERIFICATION — issue #3716 clean project-local Claude workers

Status: in progress.

## Checklist

- [ ] Exactly two required `.claude/agents` projections exist.
- [ ] Both are complete generated files derived from the canonical source.
- [ ] YAML frontmatter has required `name` and `description`, explicit minimum tools, and
  `permissionMode: default`.
- [ ] Neither worker grants `Agent`, `Skill`, MCP, an orchestration persona, or an ambient-agent
  escape path.
- [ ] Mutating a generated file to grant `Agent` fails the real drift check; sync restores it.
- [ ] Installed Claude CLI smoke discovers both project workers by name in this trusted checkout.
- [ ] Smoke uses isolated ambient user/plugin fixtures and demonstrates that delegation is blocked.
- [ ] Documentation records project discovery, precedence, tool allowlisting, and the managed/CLI
  override caveat using https://code.claude.com/docs/en/sub-agents.
- [ ] No global Claude, plugin, `.codex`, `.pi`, legacy-role, or connector path changed.
- [ ] Focused tests, canonical check, local gates, no-mistakes child gates, and sub-PR CI are green.
