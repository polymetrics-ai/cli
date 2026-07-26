# PR 528 Website Checks Repair

## Scope

Fix the failing `Website checks` e2e failure in
`website/tests/e2e/blog-annotations.spec.ts` without dropping the existing
no-mistakes fix commits on `fm/cli-release-and-connector-issues-r1`.

## Required Workflow

- GSD command attempted: `scripts/gsd prompt programming-loop init --phase pr-528-website-checks --dry-run`
- Adapter result: unavailable in this checkout (`unknown GSD command: programming-loop`)
- Fallback: manual GSD loop, after `scripts/gsd doctor` passed
- Required skills loaded: `gsd-programming-loop`, `no-mistakes`,
  `e2e-testing-patterns`, `vercel-react-best-practices`,
  `vercel-composition-patterns`
- `no-mistakes axi` custody check: local checkout reports repo not initialized;
  no branch reset, rebase, merge, push, or global tool changes are part of this
  repair.

## Diagnosis

The CI failure waits for `[data-author-preview="visible"]` after a commenter
opts into public profile details. The profile dialog allowed the visibility
checkbox to be edited while its initial `/api/profile` request was still
pending. If the test checked the box before that GET resolved, the late GET
reset `visible` back to `false`; Save then persisted a private profile, so the
later author popover rendered the private preview instead of
`data-author-preview="visible"`.

## Plan

1. Add a deterministic mocked e2e regression in `blog-smoke.spec.ts` for the
   profile-settings loading race.
2. Update `ProfileSettingsDialog` so initial settings load gates visibility
   edits and stale load completions after close are ignored.
3. Run targeted Playwright coverage for the mocked spec, plus typecheck where
   feasible.
