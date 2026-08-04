# Google Calendar documented-operation parity — resume R1

## Scope and ownership

- Branch: `fm/cli-google-calendar-parity-resume-r1`, rebased from PR #3554 onto `origin/main` at `36b431cf1`.
- Own `internal/connectors/defs/google-calendar/**`, `internal/connectors/hooks/google-calendar/**`, this phase's artifacts, and generated Google Calendar catalogs/manuals/website data. The recovered PR's minimal `native/nativeset` promotion wiring plus the conformance fixture sentinel are included because the legacy native registration otherwise shadows the bundle and fixture replay would otherwise attempt a token exchange.
- Do not change the engine, schema, validator, hook generator, or any other connector. Those paths are owned by parallel foundation lanes.

## GSD and skills

- `scripts/gsd doctor` passed. `scripts/gsd prompt programming-loop init --phase google-calendar-parity-wave03 --dry-run` returned `unknown GSD command: programming-loop`; this is the documented adapter-unavailable case, so the manual GSD workflow is in use.
- Manual lifecycle evidence: this plan, `SPEC.md`, `TEST-PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `PROMPTS.md`, and `RUN-STATE.json`.
- Skills loaded: `gsd-programming-loop`; `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.

## Provider baseline

Google's current Calendar v3 Discovery document (`revision=20260731`) lists 38 operations: 11 GET, 15 POST, 4 PATCH, 4 PUT, and 4 DELETE. The saved machine-readable research matrix is the authoritative field-level evidence for this resume.

## TDD slices

1. **Red validator evidence:** preserve the current three failures that prove `freebusy query` does not mark its three required body inputs (`timeMin`, `timeMax`, `items[].id`) as required CLI flags.
2. **Reachable read surface:** mark those flags required, keep the 11 fixture-backed GET streams and the bounded `freebusy query` direct read implemented, then run surface sync and focused conformance/CLI/runtime-preflight checks.
3. **Honest mutation ledger:** retire the recovered checkpoint's 26 executable `rest_write` actions and commands. Each documented mutation becomes a blocked `api_surface` ledger row with the named missing foundation: `rest_write` is schema-only and has no commandrunner dispatch. Do not use `planned` as a resting state.
4. **Citations and generated surfaces:** record provider-owned evidence for every declared request field in the phase research matrix; current main has no shared machine citation schema, so preserve convention-neutral evidence and regenerate only Google Calendar outputs.
5. **Verification and handoff:** run the contract's focused gates on current main, build `pm`, verify help and representative read commands without credentials, update phase records, commit the green slice, then hand branch custody to no-mistakes.

## Safety and non-goals

- No credentialed Google calls, credential access, reverse-ETL execution, provider writes, dependencies, shared-schema changes, or main-branch merge.
- `rest_write` is not an implementation path. The 26 mutation rows remain blocked until the dedicated executor foundation exists.

## Orchestration decision

`local_critical_path`: this is one connector in one isolated worktree. No subagent was requested, and the active runtime policy prohibits proactive delegation.
