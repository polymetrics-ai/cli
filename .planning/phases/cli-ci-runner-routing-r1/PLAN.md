# Plan — CI runner routing

## GSD path

- `scripts/gsd doctor`, all five `scripts/gsd sources` commands, and
  `go run ./cmd/agentcontractgen check` passed.
- Generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` prompts are executed through the documented
  inline/manual fallback: this worker brief prohibits role spawning and the
  task is not a numbered roadmap phase.

## Scope

1. Add a reusable, GitHub-hosted selector setup job. It outputs the existing
   `polymetrics-website` label only for same-repository PRs from the two
   explicitly trusted accounts; it outputs `ubuntu-latest` otherwise.
2. Wire every existing Linux GitHub-hosted job to the selector output while
   preserving their current dependencies and conditions. Leave Windows and the
   dedicated website deployment runner untouched.
3. Add a source-level routing contract test and run it from the conventions
   workflow. Update filtered workflow triggers so a selector change exercises
   affected workflows.
4. Record a ready-to-use PR body with the runner toolchain dependency and
   hardening follow-ups. No server access, provisioning, deploy, push, PR, or
   merge is in scope.

## TDD and verification

- RED: the routing contract test must fail before the selector exists.
- GREEN: it must prove the exact structural plus explicit-author condition,
  hosted non-PR fallback, all Linux consumer jobs, and unchanged Windows / site
  deploy exceptions.
- Verify YAML syntax, the routing contract, release workflow regression,
  `go run ./cmd/agentcontractgen check`, and final diff/secret inspection.

## Skills

- Required-skill routing was reviewed. This is workflow-only work, so no Go,
  CLI, or website-design skill applies. `no-mistakes doctor` passed; its later
  shipping pipeline remains reserved for firstmate.
