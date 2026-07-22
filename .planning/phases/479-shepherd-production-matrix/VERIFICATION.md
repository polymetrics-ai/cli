# Verification

## Required gates

- [ ] Focused 17-row production matrix.
- [ ] All `.pi/extensions/shepherd/*.test.ts` tests.
- [ ] Strict TypeScript for the Shepherd production closure.
- [ ] Offline Pi extension registration and help/status RPC smoke.
- [ ] `git diff --check` and changed-path ownership audit.
- [ ] No `merge main` or equivalent mutation exposed by controller, extension, or model tools.
- [ ] Live GitHub fixture when authenticated `gh` is available; otherwise a recorded human gate.

## Current environment

`gh auth status` reports an invalid configured credential. Production transport code and
deterministic transport tests remain in scope; live publication is blocked until the operator
restores authentication without exposing the token.
