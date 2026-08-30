# Issue #4365 discussion log

## Inputs resolved without a further product decision

1. The issue and launch brief fix the exact source operation, method, and
   provider path; there is no broad route-selection choice.
2. Current `origin/main@cf29d302c13f7fcd340d31ad6dc27872880ccf42` contains
   #4358 and #4360. The existing named route/base contract is sufficient, so
   this remains a connector-owned Sentry lane rather than a shared foundation
   change.
3. The required test matrix is happy, adversarial mismatch, and slash-join
   edge behavior, followed by a no-credential/zero-I/O real CLI proof.
4. Provider credentials and provider requests are out of scope. The proof
   stops at preflight and a local transport spy.
