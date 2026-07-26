# TDD Ledger

## Red

- Add mocked Playwright coverage that delays `/api/profile`, opens the profile
  dialog, and expects the visibility checkbox to remain disabled until the
  initial settings payload arrives. This fails before the fix because the
  checkbox is enabled while the load is pending.

## Green

- Change `ProfileSettingsDialog` so it enters a loading state on open, disables
  visibility edits until settings load, and ignores stale load completions.

## Refactor

- No broad refactor. Keep changes scoped to the popover interaction and the
  regression test.
