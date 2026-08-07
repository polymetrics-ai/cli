# GSD prompt record — Issue 3886

Generated inline through the project adapter after `scripts/gsd doctor`:

```sh
scripts/gsd sources discuss-phase
scripts/gsd sources plan-phase
scripts/gsd sources execute-phase
scripts/gsd sources verify-work
scripts/gsd sources code-review
scripts/gsd prompt discuss-phase dynamic-schema-discovery-foundation
scripts/gsd prompt plan-phase dynamic-schema-discovery-foundation --tdd
go run ./cmd/agentcontractgen check
```

The adapter's phase commands require a numbered Roadmap phase; #3886 is an
issue-scoped foundation and uses the documented inline/manual fallback. Its
discussion outcome is `CONTEXT.md`; its plan/TDD/verification outcomes are
the adjacent files in this directory. The later execute/verify/review prompts
will be generated and recorded before their corresponding gates.
