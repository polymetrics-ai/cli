# Verification — #3859 native database apply strategies

## Planned checks

- [ ] Focused engine/database/PostgreSQL/synctransport unit tests prove every
  added refusal has zero side effects and every success has a durable state.
- [ ] Race test for each changed concurrent registry/session package.
- [ ] Explicit Docker/Colima PostgreSQL dbtest run proves every strategy by
  resulting durable rows, not status alone.
- [ ] `go vet` on changed packages; `gofmt`; `go build ./cmd/pm`.
- [ ] All individual repository non-suite gates listed in `PLAN.md`.
- [ ] `git diff --check` and final scope review confirm no #4125/#4136/#4090
  change, no public generic write surface, and no credential material.
- [ ] CLI/docs/website parity reviewed and recorded as not applicable unless
  the final diff adds a user-visible surface.

## Result

Pending implementation.
