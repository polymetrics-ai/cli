# GSD plan — seven connector extraction r1

## Lifecycle and execution mode

Required GSD path resolved successfully with `scripts/gsd doctor`, `scripts/gsd sources
discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; canonical
projection validation passed with `go run ./cmd/agentcontractgen check`.

Generated prompts used:

```text
scripts/gsd prompt discuss-phase cli-sweep-seven-connector-extract-r1 --auto
scripts/gsd prompt plan-phase cli-sweep-seven-connector-extract-r1 --tdd
```

Inline/manual GSD execution is deliberate: this is a single-worker captain assignment, the
canonical parent contract forbids role spawning, and no compatible isolated GSD worker is available.
The fallback preserves the lifecycle and records its evidence here rather than weakening TDD,
verification, or review.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-documentation`, and `vercel-react-best-practices` (generated
website artifacts only; no React implementation is planned).

## Scope fence

Source commit: `c28bc75a3`.

Allowed authored-input paths:

```text
internal/connectors/defs/{workday-rest,jira,help-scout,greenhouse,chatwoot,gmail,lever-hiring}/
cmd/connectorgen/{chatwoot,gmail,greenhouse,help_scout,jira,lever_hiring,workday_rest}_*surface_test.go
```

Allowed generated outputs, only when produced by their documented generators:

```text
internal/connectors/defs/<seven>/cli_surface.json
internal/connectors/defs/operation_endpoint_ledger.json
docs/connectors/**
website/data/connectors.generated.json
website/lib/connectors*.generated.ts
website/lib/docs.generated.ts
website/content/docs/github-cli-surface.mdx (only if its generator changes it)
```

`github`, `zendesk-support`, all other connector bundles, `internal/connectors/engine/**`,
`internal/connectors/commandrunner/**`, generator production code, and dependency files are
forbidden. A needed change there is a foundation split and an immediate stop.

## TDD slices

1. **Plan checkpoint.** Commit this phase plan, ledger, run state, and verification checklist.
2. **Red acceptance tests.** Import only the seven source connector-specific `cmd/connectorgen`
   tests. Run their focused package selection against current `main` bundle inputs. Capture the
   actual failing output in `TDD-LEDGER.md` before importing any bundle input.
3. **Green bundle extraction.** Apply only the seven source bundle deltas, excluding generated
   `cli_surface.json`; regenerate command surfaces and the endpoint ledger with
   `connectorgen surface-sync`. Do not copy or hand-merge generated output.
4. **Green docs/data regeneration.** Build `pm`, regenerate connector manuals/skills using its
   documented docs command, then run `website`'s `gen:website-data`. Audit the generated diff; it
   must describe only the seven connectors plus shared catalog/index projections.
5. **Executable verification.** Run focused surface tests, connector validation, commandrunner
   preflight, the real-binary reachability sweep, help/docs/website checks, and generated-drift
   checks. Record exact counts and results.
6. **GSD verification and review.** Generate/execute the required `verify-work` and `code-review`
   prompts inline; resolve any gaps with the required gap sequence before final commit.

## Expected counts

| Connector | Implemented commands | Documented operations |
| --- | ---: | ---: |
| workday-rest | 911 | 911 |
| jira | 584 | 617 |
| help-scout | 139 | 144 |
| greenhouse | 127 | 138 |
| chatwoot | 100 | 148 |
| gmail | 63 | 79 |
| lever-hiring | 60 | 106 |

Total executable command count: **1,984**. For Jira, six `partial` commands are not counted as
implemented; they remain dispositioned in the documented-operation ledger.

## CLI/help/docs/website parity checklist

- [ ] `pm help <connector>` resolves each of the seven connector topics.
- [ ] Bare `pm <connector>` prints contextual group help and exits successfully.
- [ ] Every implemented command path renders its own help `NAME` line from the real binary.
- [ ] Connector manuals and skills are generated, not hand-edited.
- [ ] Website connector data is regenerated, and its seven records expose the new command counts.
- [ ] `surface-sync --check` reports no drift and the endpoint ledger delta is confined to the
  seven connectors.

## Commit and push checkpoints

Commit/push after plan, observed-red test slice, and verified green implementation slice. Never
push `main` or create a merge. The final handoff will include the required PR truthfulness text.

