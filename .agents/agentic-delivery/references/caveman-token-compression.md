# Caveman Token Compression

## Finding

`caveman` is a communication skill, not a code-generation or reasoning model. It should reduce
orchestration token usage by compressing status, prompts, and handoffs, while preserving exact
technical content.

Use it as a default for long-running orchestrators and worker handoffs. Do not use it to compress
warnings, approval gates, code, commands, test output, or security-sensitive text past clarity.

## Why It Should Not Harm Output

- It changes phrasing, not workflow order or validation gates.
- It keeps exact identifiers, commands, paths, code, and failure strings.
- It has an auto-clarity exception for safety and multi-step sequences.
- It should be used mainly in coordinator-to-worker and worker-to-coordinator communication, where
  compact structured status is better than prose.

## Runtime Discovery Notes

- Claude Code skills are loaded only when relevant or explicitly invoked, so the long body does not
  cost context until used.
- Codex skills use progressive disclosure: the session initially sees only skill name, description,
  and path, then loads `SKILL.md` when selected.
- OpenCode discovers project `.agents/skills/<name>/SKILL.md`, which makes this repo-local skill
  portable without relying on a user-global Claude install.

## Sources

- Local source skill: `/Users/karthiksivadas/.claude/skills/caveman/SKILL.md`
- Codex skills: https://developers.openai.com/codex/skills
- Claude Code skills: https://docs.anthropic.com/en/docs/claude-code/skills
- OpenCode skills: https://opencode.ai/docs/skills/
