# Verification — Issue #475

Cycle 7 verification is pending. Planning is frozen against candidate
`a3cd85a5d0871dd1c4c99dd8b30bcd609a228c45`, immutable base
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`, and the combined 11-finding stable-head campaign at
<https://github.com/polymetrics-ai/cli/pull/486#issuecomment-5037079867>. Production and tests remain
unchanged until the PLAN checkpoint is pushed, then one complete behavior RED must precede the
architectural correction. The required terminal gates are focused/full Shepherd tests, both pinned
Pi 0.80.6 strict TypeScript scopes, offline RPC, diff, base/head equality, and issue-owned scope.

The planned lifecycle verification accounts for throwing signal attachment/removal and abandoned
creation across late resolve, reject, hang, and malformed fulfillment. A successful close cannot
precede owned late work; an uncancellable create must instead reject/quarantine close within a
bound. Each row records timers, reservations, cleanup hooks, close outcome, and unhandled
rejections. The planned redaction verification spans multiline/indented/YAML continuation,
numeric/Authorization/alias/PKCS#8 forms, unmatched quote recovery, safe multiline quote
preservation, every direct/prompt/workspace/tool/handoff consumer, and deterministic total scanner
work for padded 25/50/100 KiB flows.

The remainder of this document is the retained Cycle 6 green baseline and will be superseded only
after Cycle 7 GREEN/evidence passes.

PR #486 correction Cycle 6 revalidation is complete. The independent-review findings against
`d918617a19749cd16d6bfcf3d2fee3e5146e7380` are covered by PLAN checkpoint `4f9c5a96`, committed
test-only RED checkpoint `e8422d53`, and pass at implementation head
`93314a54302e84e053ad0d6ff44371fbf1a167e0`.

## Declared Phase Equivalent

The user/parent explicitly replaced generic repository-wide verification for this sub-worker with
the issue-focused and complete Shepherd TypeScript gates below. Therefore `verificationPassed`
means every declared command here passed; it does not claim parent-level Go/connector verification.

| Gate | Status | Evidence |
|---|---|---|
| Focused AgentSession/tool-policy tests | pass | 40 passed, 0 failed, exit 0 |
| Complete `.pi/extensions/shepherd/*.test.ts` suite | pass | 177 passed, 0 failed, exit 0 |
| Deterministic scanner scale | pass | line-boundary visits equal 25,618 / 51,218 / 102,418-byte inputs |
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
# 40 passed, 0 failed

node --test .pi/extensions/shepherd/*.test.ts
# 177 passed, 0 failed
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

Cycle 6 proves that a nested sensitive mapping value retains its value-local delimiter ownership
across a newline, so its close cannot remove the outer flow-map closer and hide a later same-line
`client_secret`. It also proves that `rock-'n-roll` remains one unquoted scalar: `-` is a quote
boundary only when it is a line-local YAML sequence marker. Direct probes, serialized role prompts,
`workspace_read`, typed capability output, and handoff summary/finding fields remove every synthetic
marker and retain `[REDACTED]`; the harmless apostrophe control remains byte-identical.

Assignment decisions reuse the scanner-owned current line end instead of scanning each remaining
suffix. The typed metrics sink reports exactly 25,618 / 51,218 / 102,418 line-boundary visits for
inputs of the same sizes, establishing linear line discovery without wall-clock assertions. The
overloaded entry point also preserves `Array.map(redactSensitiveText)` callback behavior by
ignoring its numeric index. Every earlier lifecycle and redaction regression remains green.

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
