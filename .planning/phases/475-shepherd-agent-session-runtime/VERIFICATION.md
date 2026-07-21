# Verification — Issue #475

> Superseded pending Cycle 5 revalidation. Exact-head review at
> `e41f075a9b3bfb01d410296712740b54f943ba71` found rejected-reservation timer ownership and
> line/flow lexical-state failures. Cycle 4 results below remain historical evidence only.

PR #486 correction Cycle 4 revalidation is complete. The independent-review findings against
`b4061d4e1a1545b0c8810b14b510cf048385a567` are covered by PLAN checkpoint `190b0ec7`, committed
test-only RED checkpoint `21535513`, and pass at implementation head
`01b42ae168176956d864ff10f40d1c981f37ac04`.

## Declared Phase Equivalent

The user/parent explicitly replaced generic repository-wide verification for this sub-worker with
the issue-focused and complete Shepherd TypeScript gates below. Therefore `verificationPassed`
means every declared command here passed; it does not claim parent-level Go/connector verification.

| Gate | Status | Evidence |
|---|---|---|
| Focused AgentSession/tool-policy tests | pass | 31 passed, 0 failed, exit 0 |
| Complete `.pi/extensions/shepherd/*.test.ts` suite | pass | 168 passed, 0 failed, exit 0 |
| Strict no-emit TypeScript against installed Pi 0.80.6 types | pass | owned production/tests plus all Shepherd production `.ts`; exit 0 |
| Supported offline Pi extension/RPC smoke | pass | explicit 0.80.6 binary, `PI_OFFLINE=1`, RPC `get_commands`; `pm-shepherd` registered, exit 0 |
| `git diff --check` | pass | exit 0 |
| Owned-scope diff check | pass | every base diff path matches issue #475 production/test/phase scope |

## Explicit Boundary

Not run in this lane by policy: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`,
connector certification, and all other repository-wide Go/connector gates. These remain parent
integration and GitHub CI responsibilities. This is an intentional verification partition, not a
skip, pass, or failure.

## Model / Dependency / Safety Baseline

- Installed Pi: `0.80.6`.
- Required route: `openai-codex/gpt-5.6-sol`; `high` implementation/correction, `xhigh` all other
  roles.
- New dependencies: none permitted or planned.
- Live auth/model/network/credential checks: not run and not needed.

## Commands And Results

```bash
node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
  .pi/extensions/shepherd/tool-policy.test.ts
# 31 passed, 0 failed

node --test .pi/extensions/shepherd/*.test.ts
# 168 passed, 0 failed
```

TypeScript used the already-installed TypeScript 5.9.3 compiler and explicit Pi 0.80.6 package/type
roots with:

```text
--noEmit --strict --target ES2024 --module NodeNext --moduleResolution NodeNext
--allowImportingTsExtensions --skipLibCheck
```

It passed both the owned production/test input list and every non-test
`.pi/extensions/shepherd/*.ts` production file. `--skipLibCheck` suppresses unrelated declaration
issues inside the globally installed SDK dependency graph; all Shepherd inputs remain under
`--strict`, and the new runtime imports Pi 0.80.6 `AgentSessionEvent` and
`CreateAgentSessionOptions` directly.

The supported smoke used the explicit pinned binary:

```bash
printf '%s\n' '{"type":"get_commands"}' |
  PI_OFFLINE=1 /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/pi \
    --mode rpc --no-session \
    --no-extensions --no-skills --no-prompt-templates --no-context-files \
    --extension .pi/extensions/shepherd/index.ts
```

The explicit binary reports `0.80.6`; RPC `get_commands` returned success with the `pm-shepherd`
command registered.

Cycle 4 proves the four foreground cleanup rows independently: a session created during cleanup
grace or claimed before cancellation, each with never-settling `abort()` or `waitForIdle()`. Every
run settles within the test bound, quarantines and rejects subsequent dispatch before another
prompt, forces coalesced disposal exactly once, and produces no unhandled rejection. Idle wait may
be skipped only after abort exceeds its own independent bound; disposal remains unconditional.

The scanner proves unquoted YAML flow-map credentials and spaced line-start `client_secret`
assignments through direct probes, serialized role prompts, typed tool output, and handoff
summary/finding fields. Synthetic markers are absent, `[REDACTED]` is present, and non-assignment
prose remains byte-for-byte unchanged. Targeted adversarial REDs also prove that apostrophes in
ordinary prose cannot open quote state and preserved ambiguous prose cannot hide a nested flow map.

The immutable-base check retained
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; every changed path is an issue-owned Shepherd
production/test file or one of the phase's durable planning artifacts. Local HEAD and the local
remote-tracking branch were identical after the GREEN push.

## Deviations

- `manual_gsd_fallback`: the healthy repo adapter has no `programming-loop` registry entry.
- Ambient PATH drift: default `pi` resolved to 0.80.6 initially and 0.80.10 at final verification.
  All authoritative pinned checks use the unchanged explicit 0.80.6 binary/package/types.
- No Shepherd prompt asset file was needed beyond the owned typed `role-prompts.ts` builder; it
  generates the trusted role prompt and schema envelope without widening the allowed scope.
