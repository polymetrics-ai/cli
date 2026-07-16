# Phase 421 Summary

Status: planning artifacts created; red tests pending; no production code edited yet.

## Current state

- Worker branch: `refactor/421-connections-native-cobra`.
- GSD adapter doctor passed; `programming-loop` prompt command missing, so manual GSD fallback recorded.
- Required reading and skills loaded. Repo-specific `.pi/skills/go-implementation/SKILL.md` is missing; global Go skills loaded.
- Scope limited to `connections` native Cobra parser/handler/tests and directly applicable parity artifacts.

## Planned delivery

- Replace top-level `pm connections` legacy Cobra wrapper with native Cobra subtree.
- Add native `connections create` and `connections list` actions with declared flags preserving legacy repeated, bare bool, unknown-flag, extra-arg, and global late-flag compatibility.
- Remove `connections` namespace `parseFlags` call site after native adaptation.
- Preserve docs-map help and golden stdout/stderr/exit identity; no docs/website generated changes expected.
- Preserve connection-name completion seam for Phase 15 without implementing completion.

## Verification

Pending. Planned gates are in `VERIFICATION.md`.

## Safety

No secrets. No credentialed checks. No runtime services. No dependency changes. No parent/shared orchestration edits. No merge.
