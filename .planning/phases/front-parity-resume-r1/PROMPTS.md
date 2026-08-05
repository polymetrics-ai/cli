# Manual GSD trace — Front parity resume

## Adapter evidence

- `scripts/gsd doctor`: passed.
- `scripts/gsd prompt programming-loop init --phase front-parity-resume-r1 --dry-run`:
  `scripts/gsd: unknown GSD command: programming-loop`.
- `node scripts/programming-loop.mjs --help`: failed because the script is absent.

The adapter cannot activate the required programming-loop command. This phase therefore follows
the manual GSD lifecycle: PLAN → red evidence → minimal bundle edits → focused verification →
phase summary/verification state. This is a documented adapter fallback, not a quality-gate waiver.
