# Agent Assignment — Cycle 9 Seam Mapper

## Role

Read-only explorer for the consolidated Cycle 9 correction.

## Frozen Input

- candidate: `0cdcda7e049b7ecfa2fdc52027c66c5de161f2c8`
- immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
- review ledgers: `/tmp/475-REVIEW-CYCLE8-1.md`, `/tmp/475-REVIEW-CYCLE8-2.md`

## Assignment

Map every accepted review invariant to exact functions and behavior-level RED seams in:

- `.pi/extensions/shepherd/agent-session-runtime.ts`
- `.pi/extensions/shepherd/tool-policy.ts`
- `.pi/extensions/shepherd/role-prompts.ts`
- matching tests

Call out impossible or unsafe approaches, especially hostile Proxy traversal, mutable SDK-owned
objects, cleanup deadlines, and Pi 0.80.6 custom-tool types. Do not edit files, commit, access the
network, run broad verification, or inspect out-of-scope parent-owned implementation.

## Handoff Contract

Return a compact function map, a deduplicated RED matrix, implementation ordering, type-resolution
constraints, and unresolved risks. The issue worker decides and implements all changes.
