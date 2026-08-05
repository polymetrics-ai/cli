# VERIFICATION — issue #3721 Codex project-local workers

Status: in progress.

## Checklist

- [ ] Canonical source expresses Codex standalone TOML and isolation facts.
- [ ] Both target TOML files are generated, parse, and contain all official required fields.
- [ ] A test first fails because the worker can reach an ambient agent, then passes with
  `agents.enabled = false`.
- [ ] Drift check rejects a changed Codex worker configuration.
- [ ] Trust requirement and untrusted fail-closed behavior are documented without inventing a
  trust command.
- [ ] Same-filename user/project collision behavior is explicitly treated as undocumented.
- [ ] No global `~/.codex` file is touched.
- [ ] Focused tests, generation, static checks, no-mistakes child gates, and required review route
  are recorded.
