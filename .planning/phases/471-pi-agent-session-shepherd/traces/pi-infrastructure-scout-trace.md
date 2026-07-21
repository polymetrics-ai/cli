# Pi Infrastructure Scout Trace

## Prompt reference

Read-only inspection of `origin/main` Pi extension discovery, dependency/test infrastructure, and
the smallest standalone file/verification scope for issue #471.

## Files inspected

- `.pi/settings.json`, `.pi/extensions/gsd/**`, and `.pi/extensions/pi-sub-agent/**`;
- root, `build/agent`, and website package/test configuration;
- main-branch Makefile, CI verification workflow, and Shepherd shell tests;
- installed Pi 0.80.6 RPC/extension-loading behavior.

## Actions and commands

Git tree/source inspection, module-resolution probes, and offline Pi RPC command-discovery probes.
No files, refs, package state, auth state, model calls, network calls, or GitHub records changed.

## Findings

- Trusted project `.pi/extensions/*/index.ts` files are auto-discovered. Existing settings entries
  are additive, not an allowlist, so issue #471 does not need `.pi/settings.json` changes.
- Main supports issue #471 without PR #438; route model/thinking inside the SDK adapter.
- There is no root Node package, lock, TypeScript config, or Pi test runner. Node 24 can execute
  erasable TypeScript tests directly. No package or dependency should be added.
- Standalone Node cannot resolve the globally installed Pi runtime package. Keep that runtime
  import in `index.ts`; inject a structural SDK interface into tested modules.
- `pi --list-extensions` is not valid in Pi 0.80.6. Offline RPC `get_commands` with an explicit
  `-e` extension path proves registration without a model/auth/network call.
- Main CI does not run Pi tests. Focused Node and Pi smoke evidence must be recorded locally while
  root `make verify` remains the compatibility gate.

## Handoff

Use `.pi/extensions/shepherd/**`, update `.pi/README.md`, avoid settings/Makefile/CI/package edits,
and run the exact offline RPC registration smoke recorded in `PLAN.md`.
