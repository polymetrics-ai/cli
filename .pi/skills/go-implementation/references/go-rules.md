# Production Go rules (cited)

Deduplicated rules for the pm CLI monolith + declarative JSON connector engine. Cite rule numbers
in findings/dispositions. Sources: Uber Go Style Guide, Google Go Style (Decisions/Best Practices),
Go team (Effective Go, Code Review Comments), Dave Cheney, Ardan Labs, Alex Edwards, clig.dev.

## Errors

1. Wrap with `%w` (not `%v`) when the caller might unwrap; `%v` only to hide details across a
   boundary. — Uber; Google BP
2. Handle each error exactly once — wrap-and-propagate OR log-and-handle, never both. — Uber; Cheney
3. Prefer opaque errors; assert behavior, not type. Sentinels/typed errors couple packages. — Cheney
4. When matchable errors are genuinely needed, export `Err*` vars and support `errors.Is/As`
   deliberately. — Uber
5. Error strings: lowercase, no trailing punctuation (they compose into chains). — CRC; Google
6. Never string-match `err.Error()` for control flow. — Cheney
7. Error is the last return; never discard with `_` unless deliberate and obvious. — Google; EG
8. Indent the error path; happy path stays unindented (early return). — CRC; EG
9. Eliminate repetitive `if err != nil` via design (scanner/accumulator types), not copy-paste. — Cheney

## Interfaces & structs

10. Consumers define interfaces; keep them 1–3 methods. — Google; CRC
11. Return concrete types, accept interfaces. — Google
12. No interface until a second implementation or a real test-double need exists. — Google
13. Compile-time compliance checks: `var _ Iface = (*Impl)(nil)`. — Uber; EG
14. Exported struct literals use named fields; omit zero-value fields. — Uber
15. Design useful zero values (`sync.Mutex`, `bytes.Buffer` style) — kills "forgot New()" bugs,
    especially for unmarshal-built engine structs. — EG; Cheney
16. Pointer receivers for mutation/mutex/large; value for small immutable — consistent per type. — CRC

## Concurrency & context

17. Never start a goroutine without knowing when and how it stops. — Uber; Cheney
18. Forgotten-sender guard: buffer size 1 when the receiver may stop listening (context timeout),
    or the sender leaks forever. — Ardan Labs
19. No goroutines from `init()`; expose constructor + `Close()/Stop()`. — Uber
20. Prefer synchronous APIs; the caller owns concurrency decisions. — Google BP; Cheney
21. `context.Context` is always the first param; never a custom context type. — Google
22. Share memory by communicating (channels over shared pointers). — EG
23. Declare channel direction (`chan<-`, `<-chan`) in signatures. — Google BP
24. Never `t.Fatal` from a non-main test goroutine. — Google BP

## Naming & package layout

25. Package names: short, lowercase, by-function not by-contents; no `util`/`common`/`helper`. —
    Google; Ardan Labs
26. Don't repeat the package name in exported identifiers. — CRC
27. Identifier length scales with scope distance. — Google; Cheney
28. Initialisms uniformly cased (`URL`, `ID`, `appID`). — CRC; Google
29. No `Get` prefixes; no type-noise in names. — EG; Google BP
30. Prefix unexported package-level globals with `_`. — Uber
31. No mutable package-level state; inject dependencies — critical for the bundle registry
    (reload + test isolation). — Uber
32. One package comment, above the `package` clause, in one file. — Google

## Testing

33. Test helpers return results (`cmp.Diff`); they don't assert. — Google BP
34. `t.Fatal` for setup failures; `t.Error` for assertions. — Google BP
35. Prefer real transports (`httptest`, subprocess CLI runs, fixture bundles) over hand-rolled
    mocks. — Google BP
36. Test-double packages get a `test` suffix (`connectorstest`). — Google BP

## JSON / serialization

37. nil vs empty slice is an API contract: nil marshals to `null`; use `[]T{}` where consumers
    expect arrays (bundle JSON, `--json` output, generated website data). — Alex Edwards
38. `omitempty` ignores zero-valued nested structs — optional sub-objects must be pointers. — Alex Edwards
39. Malformed struct tags are silently ignored on decode (zero value, no error) — lint/test any
    hand-written tag in defs schemas. — Alex Edwards

## CLI ergonomics

40. stdout = machine/pipeable; stderr = logs/progress/errors. Never break `pm ... | jq`. — clig.dev
41. Exit 0 only on success; distinct non-zero codes for branchable failure classes. — clig.dev
42. `--json` alongside human output, kept in sync. — clig.dev; repo agent rules
43. Bare namespace → contextual help, exit 0; invalid action → usage error. — clig.dev; repo parity contract
44. Named flags over positional args for anything non-obvious; short + long forms. — clig.dev
45. Destructive ops require confirmation or explicit `--force`/`--yes` — matches the repo's
    plan→preview→approval→execute reverse-ETL rule. — clig.dev
46. Rewrite raw internal errors into actionable guidance at the CLI boundary. — clig.dev

## Sources

- https://github.com/uber-go/guide/blob/master/style.md (Uber)
- https://google.github.io/styleguide/go/decisions · /best-practices (Google)
- https://go.dev/wiki/CodeReviewComments (CRC) · https://go.dev/doc/effective_go (EG)
- https://dave.cheney.net/practical-go/presentations/qcon-china.html
- https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
- https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
- https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html
- https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html
- https://www.alexedwards.net/blog/json-surprises-and-gotchas
- https://clig.dev
