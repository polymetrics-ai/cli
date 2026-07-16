# Issue 446: Brand And License Boundaries

Issue: #446
Branch: `chore/446-brand-license`
Base: `origin/main` at `c3e62448`

## Objective

Restore one shared website PM mark with a static `P` and a second position that
alternates between `M` and an underscore cursor, then
replace the repository's Elastic License 2.0 policy with an explicit mixed
license boundary: `AGPL-3.0-only` by default and `MIT` for
`internal/connectors/defs/**`.

## GSD Mode

Manual fallback. `scripts/gsd doctor` passes, but
`scripts/gsd prompt programming-loop ...` exits with `unknown GSD command:
programming-loop`. This phase follows the programming loop manually:

1. Plan the slice and its ownership/legal gate.
2. Add red contract tests before production edits.
3. Implement the minimum shared logo and license map.
4. Refactor only after focused tests pass.
5. Run focused, website, repository, and visual verification.
6. Commit and push coherent green checkpoints.
7. Open a human-gated PR and run automated review routing.

## Evidence And Decisions

- Commit `605b006e` (PR #29) deleted
  `website/components/brand/pm-logo-mark.tsx`.
- Current `main` duplicates `PM_` markup in the navbar, sidebar, and footer.
- The README profile card retains the intended P/M animation language.
- `git shortlog -sn origin/main` reports one human author plus Dependabot. This
  reduces provenance ambiguity but does not replace owner/legal review.
- Follow Grafana's documented pattern: a default root license, a path-specific
  secondary license, and a concise `LICENSING.md` map using SPDX identifiers.
- Use `AGPL-3.0-only`, not `AGPL-3.0-or-later`, so later license versions are
  not adopted automatically.
- Do not grant a commercial license in this repository. Separate commercial
  terms may be offered only by a separate agreement from the rights holder.

## Skills Used

- `gsd-programming-loop`
- `vercel-react-best-practices`
- `vercel-composition-patterns`
- Repository website design conventions and existing PM brand assets

The routed `frontend-design` and `web-design-guidelines` skills are not
available in this runtime. This limitation is recorded rather than silently
skipping the design review.

## Scope

- Add a reusable, accessible PM SVG component.
- Keep `P` static and alternate the second position between `M` and `_`.
- Keep `PM` stable when reduced motion is requested.
- Use the component in navbar, home sidebar, and site footer.
- Replace root `LICENSE` with the unmodified official AGPL v3 text.
- Add `internal/connectors/defs/LICENSE` with the MIT license text.
- Add `LICENSING.md` with default, exception, third-party, contribution, and
  trademark boundaries.
- Reconcile README, NOTICE, CONTRIBUTING, review rubric, homepage, FAQ,
  sidebar, and footer copy.
- Add focused regression tests.

## Non-Goals

- Production deployment or merge to `main`.
- Connector runtime or generated connector data changes.
- A CLA/DCO rollout.
- A commercial license grant.
- Relicensing third-party material without the required rights.
- New dependencies.

## Red-Green Plan

1. Red: logo contract expects a shared component, complementary `M`/`_` states,
   and all three consumers.
2. Red: licensing contract expects AGPL root, MIT defs, a path map, and no
   stale Elastic License declarations in maintained repository/website copy.
3. Green: restore and integrate the PM logo component.
4. Green: add license texts/map and update all identified copy.
5. Refactor: centralize only the brand mark; keep legal wording explicit at
   each public surface.

## Verification

Focused:

```bash
cd website
npm run test:unit -- tests/brand-license-contract.test.ts
npm run typecheck
npm run build
```

Repository:

```bash
go test ./...
go vet ./...
go build ./cmd/pm
make verify
```

Visual:

- Desktop and mobile screenshot of the navbar/sidebar/footer mark.
- Reduced-motion check showing a stable `PM` mark.
- Confirm `M` and `_` occupy the same stable position without clipping or
  layout shift.

## Human Gates

- Repository-owner/legal approval of the license change.
- Any commercial-license language beyond a non-granting contact statement.
- Production deployment.
- Merge to `main`.

## Sources

- GNU AGPL v3: https://www.gnu.org/licenses/agpl-3.0.en.html
- SPDX AGPL-3.0-only: https://spdx.org/licenses/AGPL-3.0-only.html
- SPDX MIT: https://spdx.org/licenses/MIT.html
- Grafana licensing map: https://github.com/grafana/grafana/blob/main/LICENSING.md
