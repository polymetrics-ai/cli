# Manual code review — CI toolchain and integration branch guard

## Scope review

Clean. The production diff is limited to the requested workflow pins, `go.mod`'s matching toolchain directive, and the convention guard. The rest is required delivery evidence.

## Security and guardrail review

Clean. `govulncheck` still runs unchanged; only its actual toolchain changes. No allow-list, suppression, skip, continuation, or non-blocking setting was introduced.

The branch-name change is a distinct anchored expression:

```bash
^integration/[1-9][0-9]*-[a-z0-9][a-z0-9._-]*$
```

It requires a positive decimal issue number and preserves the previous lower-case slug grammar. The pre-existing exception `case` and conventional type/description expression are untouched, so their accepted/rejected spaces do not broaden.

## Result

No actionable findings. Automated Claude review remains pending PR creation.
