# Issue 4325 — Batch 1 Connector Repair Plan

## Task Delivery Header

- Issue: Refs #4325 — restore batch 1 independent gate
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main` from
  `fm/cli-batch1-repair-r1`, with the independent credential-free Gate B rerun
  returning GO and all repository checks green.
- Working branch: fm/cli-batch1-repair-r1
- Task: Repair the ten batch-1 source bundles and their derived command,
  operation, and evidence surfaces without credentials or reduced quality gates.
  Each connector is a sequential ownership slice; any missing shared runtime
  capability is split into a foundation issue instead of being shimmed locally.
- Verification: source-import/check, `connectorgen validate`, `surface-sync
  --check`, real commandrunner preflight, focused regression tests, built-binary
  credential-boundary probes, GitHub source-lock checksum assertions, full
  `make verify`, and an independent Gate B rerun.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Current provider locks map exactly to declared operation identities | live | A fresh credential-free source import accepts the pinned bytes and its descriptor identity set matches the provider document; a stale or missing operation fails the check. |
| Enabled command surface is executable | live | Real `commandrunner.Preflight` accepts every implemented command and a built binary dispatches selected commands to `missing --credential`, not `unknown command`. |
| Disabled evidence is truthful | live | The independent ledger audit finds zero `requires-elevated-scope` reasons and every remaining `foundation-gap` contains a current refusing `file:line`. |
| GitHub parity remains immutable | live | The GitHub lock and descriptor file byte counts and SHA-256 equal the captain-specified values. |
| No provider is falsely certified | live | `pm connectors inspect <connector> --json` retains `live_certification: pending` for every batch connector. |

## Required skills and workflow

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation`.
- GSD prompts resolved and executed inline: `discuss-phase 4325`,
  `plan-phase 4325 --tdd`; `execute-phase`, `verify-work`, and `code-review`
  will be resolved and recorded at their respective gates. Inline execution is
  required because compatible isolated runtime agents are unavailable and the
  canonical contract forbids role spawning.
- CLI parity: this changes generated connector command surfaces. Verify the
  bare `pm connectors` namespace, `pm connectors inspect <name> --json`, one
  changed connector command `--help`, generated docs checks, and discovery
  metadata. Hand-authored CLI/website pages are not applicable unless a
  generator reports changed output.

## Sequential TDD slices

1. **Baseline/red inventory.** Record failing source-import/check results for
   all mutable locks, zero-surface command probes, drifted provider identities,
   Jira’s untyped/read-write accounting, and Stripe’s JSON misclassification.
2. **Source integrity and identity repairs.** Refresh the six mutable sources;
   remove Bitbucket/GitLab stale rows and add Notion’s twelve current rows;
   regenerate descriptors, parameter mappings, and derived surfaces. Red is a
   source-import/check mismatch or set-difference test; green is a verified,
   exact operation identity set.
3. **Terminal surfaces.** Adopt the source-derived, bounded operations needed
   for CircleCI, Sentry, and Vercel commands. Red is a built-binary
   `unknown command`; green dispatch reaches the credential boundary and real
   runtime preflight accepts the command.
4. **Reachability truth.** Convert Jira enabled direct reads to typed operation
   contracts and align its direct-write ledger with executable commands; fix
   Docker Hub metadata/citations and Notion’s comment-action binding. Red is
   an unbound enabled row or contradictory state; green is a typed join and
   credential-boundary probe.
5. **Classification and disabled evidence.** Reclassify Stripe file metadata
   responses as JSON direct reads and regenerate all dispositions so forbidden
   scope reasons disappear and foundation gaps cite present code. Red is the
   independent evidence scan; green is zero forbidden reasons and complete
   citations.
6. **Gate completion.** Run structural checks, focused tests, binary probes,
   GitHub checksums, full `make verify`, then the independent Gate B rerun.
   If it finds gaps, use `plan-phase 4325 --gaps` and
   `execute-phase 4325 --gaps-only` before repeating verification.

## Commit and push checkpoints

- Commit this plan and the red baseline separately from production repairs.
- Commit each connector-complete green slice separately, without unrelated
  bundle changes.
- Run full `make verify` serially before every push; never force-push, rebase a
  published branch, or push `main`.

## Safety boundaries

- No provider credential, browser profile, runtime state, reverse-ETL execution,
  destructive provider request, new dependency, test suppression, or quality
  reduction.
- Never edit `internal/connectors/defs/github/rate_limits.json`.
- Keep GitHub lock/descriptor bytes and SHA-256 unchanged; report measured
  values rather than claims.
