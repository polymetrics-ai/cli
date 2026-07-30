# PLAN — website profile settings loading

## Scope

Preserve only the independently justified website regression fix discovered in the closed `origin/ci/release-publishing` history: the profile settings dialog must wait for `/api/profile` settings to load before enabling profile visibility/url edits or Save.

## Branch comparison evidence

- Current base: `origin/main` at `902fd9403`.
- Preserved branch inspected: `origin/ci/release-publishing` at `53158535f`.
- Relevant historical commit: `56d1bea2a no-mistakes: apply CI fixes` added the profile-settings loading guard and Playwright regression test.
- Later historical commit: `29cd7b1ab ci(release): prepare pm v0.1.0 binary release` removed that guard/test while adding release-coupled work, so final `origin/ci/release-publishing` no longer contains the fix.
- Current `origin/main` still lacks the loading guard in `website/components/auth/profile-settings-dialog.tsx` and lacks the regression in `website/tests/e2e/blog-smoke.spec.ts`.

## Explicit exclusions

Do not retain release-publishing branch changes for:

- release-triggered website dispatch / `website-release` jobs;
- release tag inputs, PM binary release coupling, release docs, or deploy docs;
- connector generated catalog differences;
- connector code, issue guards, or release workflow changes.

## Implementation slices

1. Red test: add a focused Playwright regression that opens the blog profile dialog, stalls `GET /api/profile`, asserts the visibility checkbox is disabled while settings are unresolved, then releases the response and proves a user can enable and check the visibility control.
2. Green code: reset `settings` when the dialog opens, derive `loading` from unresolved settings, ignore stale `/api/profile` responses after close/unmount, and disable visibility/url controls plus Save while loading or saving.
3. Verification: run the focused e2e test, website generated-data check, typecheck, unit tests, and build as locally available.

## GSD / skill evidence

- GSD adapter health: `scripts/gsd doctor` passed.
- Attempted required command: `scripts/gsd prompt programming-loop init --phase website-profile-settings-loading --dry-run` failed with `unknown GSD command: programming-loop`; used repo-local Pi prompt `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` as manual GSD programming-loop fallback.
- Required routing loaded: `.agents/agentic-delivery/references/required-skills-routing.md`.
- Website skill loaded: `vercel-react-best-practices`.
- `frontend-design` and `web-design-guidelines` were named by repo routing but are not installed in the available skills directories; no visual/API composition change is planned.

## Orchestration decision

`local_critical_path` — the task is one small website component/test slice in an already isolated disposable worktree; spawning a mutating subagent would add coordination overhead without parallelizable scope.
