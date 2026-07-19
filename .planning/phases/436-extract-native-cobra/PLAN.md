# Phase 436 Plan — hidden extract native Cobra command

Issue: polymetrics-ai/cli#436
Umbrella: #407
Parent: #397 / draft parent PR #438
Branch: `refactor/436-extract-native-cobra`
Base branch: `feat/cli-architecture-v2`
Exact starting HEAD: `eec03373dcc581c7f5c3331fe63287519b317f53`
Invocation session: `issue-436-pi-sol-high-20260719T074902Z`
Explicit invocation profile: `Sol`, `high`
Execution decision: `local_critical_path` — #436 is the assigned next serialized Phase 9 unit in an isolated worktree. Its central router scope collides with sibling namespace migrations, this session exposes no subagent tool, and the user bounded delivery to implementation/commit/push with no PR or review.

## GSD route

- `scripts/gsd doctor` and `scripts/gsd list`: passed (69 commands).
- `scripts/gsd prompt plan-phase 436 --skip-research`: generated and executed inline.
- `scripts/gsd prompt programming-loop 436`: unavailable because the adapter registry has no `programming-loop` command.
- Manual fallback: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, with these six issue-local artifacts and strict RED → GREEN → refactor evidence.

## Issue contract and required reading

The direct issue task requires the next serialized #397/#407 Phase 9 unit: preserve hidden `extract`, nativize its current command/flag/operand behavior, contain input/output paths and stop final-link escapes with temporary-root tests, preserve bare/text/JSON/positional/trailing help and literal-`--`/unknown/action/operand/global/output behavior, and remove only extract's legacy parser registration/call. No external files, services, credentials, dependencies, PR, or review.

Read: `AGENTS.md`; local #397/#407 orchestration artifacts; GSD core/adapter/manual loop; issue contract; CLI parity policy; ADR-0002; CLI Architecture v2 plan; current extract/router/golden/manual/website code; adjacent native RLM/worker commands; RLM warehouse read/write code; and `internal/safety` rooted filesystem helpers.

Loaded: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-project-layout`; `golang-documentation`; `golang-spf13-cobra`.

## Scope

- Replace only hidden `extract`'s legacy Cobra wrapper with a hidden native Cobra command and hidden positional help compatibility.
- Declare the complete current local flag surface with repeated last-value and legacy bare/assigned/space semantics: `--request`, `--sql`, `--limit`, `--provider`, `--model`, `--llm-base-url`, `--in`, `--out`, and `--spec-name`.
- Preserve ignored trailing operands and unknown flags, dependency-free heuristic routing, optional typed RLM escalation, simple read-only SQL routing, exact `ExtractResult` text/JSON semantics, context propagation, error taxonomy, stdout/stderr discipline, and invocation-global `--root`, `--json`, `--plain`, `--no-input`, and `--progress` placement/assignment behavior.
- Preserve strict first-command ownership: literal `--`, malformed/legal unknown command-head flags, invalid actions/operands, and later `extract`-looking tokens must not discover or execute an extract effect.
- Add an invocation-local extract runtime seam so focused parsing/help/error tests can observe query/analyzer routing without model, Temporal, Podman, worker, database service, or network calls.
- Validate extract RLM input/output table names as bounded local identifiers before analyzer construction/effects. Harden the shared local RLM warehouse file read/write seam only as required to ensure rooted input reads, rooted temporary output creation, and final-link-safe atomic replacement under the selected temporary warehouse root. Existing explicit table-name and result semantics remain.
- Remove only extract's entry from `cobraLegacyCommands`, extract's `parseFlags` use, and any now-dead extract-only parser adapter. Dynamic connector parsing and all other namespace parsers remain untouched.

Excluded: new extract actions/routes; SQL write support; generic model/shell/HTTP/SQL tools; RLM model implementation; worker/Temporal/Podman changes; connector definitions; credentials; optional services; broad RLM redesign; dependencies; Phase 16 dashboards; Phase 19 broad help/man work; PR/review.

## TDD slices and checkpoints

1. **Planning checkpoint** — commit/push these six artifacts before test or production edits.
2. **RED checkpoint** — add focused tests specifying:
   - hidden native extract ownership, no legacy wrapper, all current pflags/NoOpt/completion seams, no unapproved child actions, and only extract parser removal;
   - repeated/bare/assigned/space flags, ignored trailing operands, dependency-free simple/RLM route selection, exact text/JSON output, query/analyzer context, and no side effects on parse/help/error paths;
   - bare/text/JSON/long/short/positional/trailing help with hidden root discovery;
   - literal `--`, malformed/legal unknown command heads and tails, invalid actions/operands, no later action discovery, and current global forms;
   - temporary-root input/output traversal rejection, input final-link escape rejection, output temp/final-link safety, external sentinel preservation, valid in-root input/output operation, and no broad filesystem paths.
   Capture focused failure before production edits; commit/push tests.
3. **GREEN checkpoint** — add the smallest hidden native extract command, typed flags/handler/runtime seam, bounded extract-only normalization, table validation, and rooted RLM warehouse I/O needed by the RED contract. Remove only the extract legacy wrapper and extract parser call.
4. **Refactor/parity checkpoint** — run focused/repeated/race extract/RLM/safety, router/golden/full CLI, exact-start parser/output differential, runtime help, generated manual/docs/website parity, formatting, vet, build, and scope/dependency guards.
5. **Final checkpoint** — run full dependency-free verification including `make verify`; finalize all six artifacts; commit/push; no PR or review.

## CLI help/manual/website parity

Hidden status is invariant: extract remains absent from root discovery/completion lists. Direct bare and help-topic behavior becomes a contextual hidden manual per #436 acceptance, with text and JSON parity. Add/update only directly applicable `extract` manual, website agent-guide/CLI-reference references, and reviewed golden fixtures. Verify `pm help extract`, bare `pm extract`, `pm extract --help`, `-h`, positional help, trailing help, JSON help, invalid inputs, stdout/stderr, docs generation, website generation, and hidden root discovery. Phase 19 retains broad help/man work.

## Safety

All tests use `t.TempDir` project/warehouse roots, local synthetic records, injected query/analyzer fakes, or the existing hermetic fake runner. No model, Temporal, Podman, worker, network listener, database service, external file, credential, connector check, generic execution surface, destructive/admin action, production deployment, or standalone reverse ETL is permitted. Input and output names cannot select broad paths. Rooted effects must reject symlink escapes at effect time and preserve external sentinel content. No dependency or quality-gate change.
