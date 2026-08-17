# Context — CLI required-flag derivation r1

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification, P1 required REST path-parameter parity.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with recorded local gates and API-confirmed base.
- Working branch: `fm/cli-required-flag-derivation-r1`
- Task: Derive `cli_surface.json` required flags from required `operations.json` REST path parameters for every connector; regenerate artifacts; prove no optional mapped flag remains; re-run GitHub parity and verify every declared unsupported classification without silently reclassifying it.
- Verification: Focused and package tests (including `go test -timeout 20m ./cmd/connectorgen`), generator validation/checks, repository boundary and docs generators run twice for byte stability, runtime help/preflight checks, and scoped static sweeps with count reports.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Required path parameters produce required flags across bundles | live | A generated-surface sweep loads every command and its REST parameter declarations; it finds zero optional flags mapped to required path parameters. |
| Missing a required path flag is a typed pre-I/O refusal | fake | A deterministic command-runner transport records zero calls while the real validation path returns its typed usage error; live GitHub execution is unnecessary and would require a credential. |
| GitHub's 92 findings close through regeneration | live | The repository parity sweep reports zero required-path-flag findings for GitHub after generator output is regenerated. |
| Other connector impact is measured | live | The same sweep emits before/after counts grouped by connector from all definition bundles. |
| Unsupported declarations are honest | live | The verifier enumerates all 50 declared unsupported commands and checks each against its provider-surface/operation declaration, reporting supported contradictions without changing classifications. |
| Generated files are deterministic | live | A second run of each repository generator produces no diff. |

## GSD execution note

`scripts/gsd doctor`, each required `scripts/gsd sources` command, and the generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved. This direct-PR worker is executing the prompts inline because the canonical project contract forbids spawning GSD roles in this environment.

## Discussion resolution

The brief settles the material design choice: generator-owned derivation, never GitHub-specific edits or an allowlist. The remaining work is implementation and evidence collection; no product decision is pending.
