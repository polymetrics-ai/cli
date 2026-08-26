# Discussion log — #4354

The launch brief settles the material choices:

- Base on the then-current `origin/main` (`060bb7864e3419e09ab10e000bb14ac1ea3724ec`); use a standalone PR to `main`.
- Use candidate `18248d233e6abd9d7ec03075a225cf35ee2f5399` only to locate Outreach-owned artifacts. Do not import Batch 6–7 planning or other connector changes.
- The required product result is evidence integrity plus command-boundary proof, not provider execution.
- All six lanes must be stated; binary lanes may only be not applicable with source evidence.
- The active foundation owner, PR #4350, prohibits a competing schema-v3 source-evidence change. This pilot will expose that dependency precisely if it blocks verification.

No further user choice is needed: the source lock determines the inventory, lane names are fixed by the brief, and safe fixture-only preflight determines the runtime proof.
