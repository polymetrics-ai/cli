# Batch R1 vNext cutover discussion log

## 2026-09-01 restart decision record

The captain has already made every product and architecture decision needed for this continuation:

- Finish legacy cleanup before migrating any additional connector.
- Use only `source.lock.json -> canonical descriptor -> deterministic execution JSON -> existing PM-CLI runtime`.
- Delete alternate execution/admission paths; do not preserve them as fixtures, compatibility adapters, feature flags, importers, certification, or retention gates.
- Do not introduce a shared runtime foundation, recover excluded local work, create credentials, use provider I/O, merge, force push, rebase, or open another PR.
- Migrate the named remaining connectors in fixed order and push every independently green cohort normally to the established Batch R1 remote branch.

No interactive product question remains. The only operational constraint discovered at restart is the missing GSD ROADMAP/prompt prerequisite described in `CONTEXT.md`; it is handled through the documented inline/manual-GSD evidence fallback, not by changing scope or weakening TDD/review requirements.

## 2026-09-04 A1 entry-capacity review disposition

Firstmate's exact-SHA review supplies the closed product decision: correct only `BundleStore` entry-capacity accounting, add the named barrier regression, preserve every listed boundary, and obtain one new independent review before any parent publication. `--auto` is safe because no product, architecture, scope, or testing choice remains open.

Ranked diagnosis record:

1. `reserveLocked` compares only `len(cache)` to `Limits.Entries`, so a second distinct key starts while the first key is only byte-reserved. Prediction: `Entries: 1, Bytes: 2` permits two barrier-held loaders before either completion.
2. A flight's entry reservation must outlive cancellation until the loader exits. Prediction: if cancellation releases it early, the second identity can begin while the first loader is still live; if it never releases, the retry remains capacity-blocked after completion.
3. Same-key joining must remain unchanged. Prediction: the existing concurrent same-identity test still invokes one loader after count reservation is added.

The regression will use the existing package-private store state only to assert the bounded resource invariant; its acceptance behavior is distinct-identity admission/rejection/retry. No public debug API or capacity knob is added.
