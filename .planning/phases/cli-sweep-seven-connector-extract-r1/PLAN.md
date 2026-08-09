# GSD plan — seven connector extraction r1

## Lifecycle and execution mode

Required GSD path resolved successfully with `scripts/gsd doctor`, `scripts/gsd sources
discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; canonical
projection validation passed with `go run ./cmd/agentcontractgen check`.

Generated prompts used:

```text
scripts/gsd prompt discuss-phase cli-sweep-seven-connector-extract-r1 --auto
scripts/gsd prompt plan-phase cli-sweep-seven-connector-extract-r1 --tdd
scripts/gsd prompt execute-phase cli-sweep-seven-connector-extract-r1
scripts/gsd prompt verify-work cli-sweep-seven-connector-extract-r1
scripts/gsd prompt code-review cli-sweep-seven-connector-extract-r1
```

Inline/manual GSD execution is deliberate: this is a single-worker captain assignment, the
canonical parent contract forbids role spawning, and no compatible isolated GSD worker is available.
The fallback preserves the lifecycle and records its evidence here rather than weakening TDD,
verification, or review.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-documentation`, and `vercel-react-best-practices` (generated
website artifacts only; no React implementation is planned). `golang-lint` was loaded before the
mandatory review stage; `make lint` reported 0 issues.

## Scope fence

Source commit: `c28bc75a3`.

Allowed authored-input paths:

```text
internal/connectors/defs/{workday-rest,jira,help-scout,greenhouse,chatwoot,gmail,lever-hiring}/
cmd/connectorgen/{chatwoot,gmail,greenhouse,help_scout,jira,lever_hiring,workday_rest}_*surface_test.go
```

Captain-authorized shared-foundation paths, committed separately from the bundle import:

```text
internal/connectors/engine/bundle.go
internal/connectors/engine/schema/api_surface.schema.json
cmd/connectorgen/validate.go
cmd/connectorgen/validate_surface_test.go
```

The named reason is that Jira and Workday model several distinct write contracts over one documented
provider endpoint. The foundation accepts and validates plural `covered_by.writes`; it does not
import other source-branch engine changes, alter current-main REST operation parameters, or bring
`github`/`zendesk-support` tests into this extraction.

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

`github`, `zendesk-support`, all other connector bundles, `internal/connectors/engine/**` except the
four explicitly authorized foundation files above, `internal/connectors/commandrunner/**`, all other
generator production code, and dependency files are forbidden.

## TDD slices

1. **Plan checkpoint.** Commit this phase plan, ledger, run state, and verification checklist.
2. **Red acceptance tests.** Import only the seven source connector-specific `cmd/connectorgen`
   tests. Run their focused package selection against current `main` bundle inputs. Capture the
   actual failing output in `TDD-LEDGER.md` before importing any bundle input.
3. **Red/green shared foundation.** Add a focused validator test for plural write coverage, observe
   its compile-time red state against current main, then implement only the authorized engine/schema/
   validator support. Run validation across all 551 bundles with zero findings before proceeding and
   commit this foundation as its own clearly labelled change.
4. **Green bundle extraction.** Apply only the seven source bundle deltas. Import each
   connector-local `cli_surface.json` atomically because its command taxonomy contains
   non-derivable contract data; then regenerate its derivable fields and the endpoint ledger with
   `connectorgen surface-sync`. Do not field-level hand-merge generated output.
5. **Green docs/data regeneration.** Build `pm`, regenerate connector manuals/skills using its
   documented docs command, then run `website`'s `gen:website-data`. Audit the generated diff; it
   must describe only the seven connectors plus shared catalog/index projections.
6. **Executable verification.** Run focused surface tests, connector validation, commandrunner
   preflight, the real-binary reachability sweep, help/docs/website checks, and generated-drift
   checks. Record exact counts and results.
7. **GSD verification and review.** Generate/execute the required `verify-work` and `code-review`
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
