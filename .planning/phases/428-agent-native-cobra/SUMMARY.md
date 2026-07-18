# Phase 428 Summary

Status: implementation complete, accepted Medium second correction and prior High correction fully verified and pushed; no PR created.

## Accepted Medium second correction

- Invocation/session: `Second bounded correction for #428 at start af604d7178c83d21d77abedfa3c4dee29f94c089`; `issue-428-second-correction-pi-20260718T140317Z`; runtime model identity not exposed.
- Exact start/end: `af604d7178c83d21d77abedfa3c4dee29f94c089` at `20260718T140317Z`; verified implementation head `7f23901bf08fd11ac1384509396ca1a46c9d16ff`; verification end `20260718T141424Z` UTC.
- Accepted `/tmp/pm-397-rereview-428.log` Medium. A one-line agent-only tail predicate now strips clustered single-dash tokens containing lowercase `h` after exact plan/build/pull/ensure actions, preventing Cobra help interception while leaving ordinary exact `agent -h`, invalid heads, root, and other namespaces unchanged.
- RED first: all 20 `-hx`/`-xh`/`-hh`/`-xhy`/`-zzhzz` × plan/build/pull/ensure cases rendered agent help and suppressed expected actions (`0.582s`). GREEN passes exact plan output and fake runtime lookup/file/run traces (`0.561s`).
- Focused agent/router (`4.462s`), repeated (`0.579s`), adversarial (`0.584s`), race (`1.715s`), 32/32 exact base exit/stdout/stderr/fake-runtime differential, full CLI (`239.818s`), gofmt, vet, build, diff, scope, and dependency gates pass.
- Commits: `d5183180` planning, `baae3933` RED, `7f23901b` GREEN, plus final artifact checkpoint. All checkpoints pushed; no PR/external review, container/service, real image operation, dependency, credential, secret, or docs/website/generated change.

## Accepted High correction

- Correction session: `issue-428-review-fix-pi-openai-codex-gpt-5.6-sol-high-20260718T132841Z`; model/thinking `openai-codex/gpt-5.6-sol`/`high`; exact start `746b2a98b01ba1e119974e31569fc8deb06cd897`; end `20260718T134141Z`.
- Added an agent-only pre-Cobra action boundary: non-exact agent/image action heads are preserved behind `--`, so Cobra cannot discover later `image build|pull|ensure` nodes.
- Image-parent invalid tails now return the exact legacy unknown-subcommand usage error before runtime lookup. Exact `plan`, `image`, `build`, `pull`, `ensure`, agent help, and post-action unknown/help/literal behavior remain intact.
- Test-first RED (`0.587s`) exposed assigned unknown/help-like and image literal bypasses. GREEN passes 30 fake-runtime cases with zero lookups/files/runs; focused (`4.446s`), race (`1.679s`), repeated (`0.582s`), 35/35 exact base differential, full CLI (`234.335s`), gofmt, vet, build, diff, scope, and dependency gates pass.
- Correction commits: `46556e9d` planning, `5cc5fa40` RED, `7ff7debf` implementation, plus final artifact checkpoint. No containers/services, dependencies, secrets, image actions, docs/website/golden changes, PR, or external review.

## Identity

- Session: `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/428-agent-native-cobra`
- Exact start: `235233f7cfde4a24612be6b0f95fb37a412d388a`
- Verification end: `20260718T131634Z`
- Parent: #397; umbrella: #407; draft parent PR #438

## Delivered

- Registered native Cobra nodes for `agent`, `agent plan`, `agent image`, `agent image build|pull|ensure`, and hidden positional `agent help`.
- Declared repeated `StringArray --request` with bare `true`, unknown-flag tolerance, no-file completion seam, and exact spaced/assigned/repeated-last-wins behavior.
- Removed `agent` from the legacy wrappers and removed only the agent handler's `parseFlags` call.
- Preserved bare/text/JSON/long/short/positional help; deterministic plan text/JSON; invalid-action categories; assigned globals/config; action-tail `--help`/`-h`; and continuation after literal `--`.
- Added a context-aware injected image runtime and fake-only coverage for build, pull, ensure-present, ensure-pull, exact command arguments/paths/output, failures, and invocation order.
- Added bounded validation for agent request controls, project/build root, Podman binary, and image references before external lookup/execution.
- Kept worker/RLM behavior, canonical help/docs/website/goldens, dependencies, connector definitions, other namespaces, and Phase 15/19 scope unchanged.

## TDD and verification

The test-only RED failed exactly on the intentionally missing `runAgentImageAction` and `newRootCmdWithAgentImageRuntime` seams before production edits. Focused GREEN passed in `4.408s`; expanded/final focused runs passed in `4.480s`/`4.386s`. Agent/router/golden passed in `11.408s`; standalone golden in `5.816s` and `6.054s`; full CLI in `235.686s`; focused race in `1.751s`.

A 25-case base/head differential matched exact exit/stdout/stderr for help, plan flag forms, unknown/positional input, assigned globals, trailing help, literal separators, and missing/invalid image actions. Built help routes were byte-identical (`450` bytes), plan output deterministic, invalid action exit `2`, and invalid assigned boolean/unsafe request exit `3`.

Runtime dependency-free config/RLM/worker tests passed without optional services. Temporary CLI docs generation diff, temporary/tracked connector docs validation, website generation (11 pages), golden, dependency, connector-def, and scope guards were clean. gofmt, vet, build, full repository tests (`real 345.240s`; CLI `238.990s`, certify `341.079s`), and `make verify` (`real 25.853s`; smoke OK, lint `0 issues`, 547 connectors / 0 findings) passed.

## Workflow and safety

GSD doctor/list/plan-phase passed. Programming-loop is absent, so the recorded manual universal-loop fallback enforced strict TDD. Verify-work and code-review prompts generated (7137/6003 bytes) and were executed inline; local review was clean. No external review was requested.

No secret, credentialed connector, optional service, Podman/Docker command, image build/pull/publish, Temporal/PostgreSQL/Dragonfly operation, dependency, worker behavior change, PR, or merge occurred. All image behavior used injected fakes/temp roots. Invalid-action differential used executable lookup only, not execution. The required `make verify` local sample smoke followed reverse ETL plan → preview → approval → run without external writes.

## Delivery

Commits pushed:

- `a37719a7` — planning checkpoint
- `6cdaea60` — RED test checkpoint
- `4e1b20f3` — native implementation checkpoint
- `75e0d732` — validation hardening checkpoint
- final verification/handoff artifact commit

Residual risk: the agent-only pre-Cobra normalizer intentionally preserves legacy ignored help/separator tokens until Phase 19 deliberately changes focused help. Tests and the exact differential pin that boundary.
