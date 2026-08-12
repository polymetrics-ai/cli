# Issue #4072: GitHub App auth admission - Discussion Log

> **Audit trail only.** Do not use as input to planning or execution agents.
> Locked implementation decisions are in `CONTEXT.md`.

**Date:** 2026-08-12
**Phase:** issue-4072-github-app-auth-admission-r1
**Areas discussed:** recovery custody, shared-rate admission, request capability, TDD evidence

---

## Recovery and ownership

| Option | Description | Selected |
|--------|-------------|----------|
| Reuse a live exact-owner issue | Avoid creating a duplicate child | |
| Create direct #3754 child #4072 | No exact owner existed; attach and verify before code | ✓ |
| Continue exhausted #3754 lineage | Reuse 5/5 evidence | |

**User's choice:** Fresh #3754 child, lineage 0/5, from preserved
`da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`.

---

## Token-request admission

| Option | Description | Selected |
|--------|-------------|----------|
| Keep raw `http.DefaultClient` in the GitHub hook | Bypasses resolver and coordinator | |
| Expose coordinator to hooks | Couples hooks to shared-budget internals | |
| Engine-owned declared-route request capability | Uses selected policy without exposing raw coordinator | ✓ |

**User's choice:** The Sol audit's architecture-safe repair: resolver before
custom auth and a narrow engine-owned declared-route request capability.

---

## Focused-stage validation

| Option | Description | Selected |
|--------|-------------|----------|
| Real GitHub credentials/provider exchange | External state and secrets | |
| Local fake transport/coordinator only | Deterministic zero-send and lifecycle evidence | ✓ |
| Broad validation/no-mistakes now | Conflicts with Firstmate shared CPU gate | |

**User's choice:** Commit focused RED/GREEN only, then pause for Firstmate.

---

## the agent's Discretion

- Select the smallest private capability shape that preserves all existing
  custom hooks and does not add a generic user-facing HTTP writer.
- Add only the test plumbing necessary to inject a local recording transport.

## Deferred Ideas

No broader delivery action: no no-mistakes run, push, PR, merge, or parent
route selection until Firstmate releases the shared validation gate.

---

*Issue phase: issue-4072-github-app-auth-admission-r1*
*Discussion log generated: 2026-08-12*
