# Issue #4015 — GitHub declared-command parity

## Outcome

Evaluate all 50 GitHub commands in the supplied parity inventory and implement every command that
can be expressed through the repository's typed connector runtime. Preserve each remaining command
with concrete provider or product-boundary evidence. The verdict inventory must total exactly 50.

## Locked scope

- Target connector: `github` only.
- Source inventory:
  `/Users/karthiksivadas/karthik-agent-workspace/data/parity/unimplemented-50.json` — 27
  `unsupported_api` commands and 23 `unsupported_local` commands.
- Working branch: `fm/cli-parity-implement-50`.
- Pull-request base: `integration/4015-mvp-flat-r1`; final landing path is
  `integration/4015-mvp-flat-r1 → main` under the existing human gate.
- The launch brief is autonomous and supplies the implementation choices normally collected in an
  interactive discussion. This phase therefore uses `discuss-phase --auto`.
- Fixed provider operations must use reviewed, provider-documented REST or GraphQL declarations.
  An empty `api_surface` is not provider evidence.
- `api`, auth-token display, generic local shelling, and other unrestricted escape hatches remain
  prohibited by the connector canon even if the upstream `gh` CLI exposes them.
- Local-only commands may be implemented only through an existing typed, dependency-free local
  workflow with bounded arguments and testable behavior. Do not add a generic process runner.
- No declared command is removed. A command that cannot be implemented retains its declaration and
  receives a concrete verdict/evidence record.

## Required behavior

- Derive command flags and `api_surface` from fixed `operations.json` declarations through
  `connectorgen surface-sync`; do not hand-author generated fields or opaque provider cursors.
- Implemented direct reads return one honest page and the shared direct-read page context.
- Implemented mutations retain plan → preview → approval → execute and fixed response bounds.
- GraphQL mutations use fixed reviewed documents and typed variables, never caller-supplied query
  text.
- Live verification uses only the disposable `Polymetrics-Cert` identity and
  `pm-cert-3993-20260810-wz0fru` repository. Credentials are read from Keychain into an environment
  variable at point of use and never written, printed, summarized, logged, or placed in argv.
- Every disposable resource is prefixed `pm-cert-`, removed after the proof, and independently read
  back as absent; read commands assert returned values or counts, not merely exit status.

## Delivery choices

- Work in coherent typed families, committing and pushing each green slice.
- Start with metadata-only/runtime preflight tests that fail for the 50 empty surfaces, then add
  endpoint-family tests and live proof per implemented command.
- Generate manuals and website connector references using existing repository generators. Record
  any parity surface that is not applicable rather than silently omitting it.
- Execute the GSD workflow inline because this canonical single-worker lane forbids role spawning.

## Safety boundaries

- No third-party repositories, real people, public org-visible fixtures, token-scope changes,
  purchases above USD 2 per operation, new dependencies, production deploys, or merge to `main`.
- No ambient `gh` login and no secret-bearing output command.
- Destructive provider actions are limited to explicitly authorized cleanup of this task's
  disposable fixtures, followed by independent absence checks.

