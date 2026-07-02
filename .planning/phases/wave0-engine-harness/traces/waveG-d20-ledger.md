# D-20 — migration docs + agent I/O schemas (docs-only, wave G)

Phase: wave0-engine-harness · Wave: G · Executor: gsd-loop-backend (sonnet), docs-only, no TDD pair.

## Scope

Wrote the three artifacts that gate every Pass A/Pass B fan-out migration agent and reviewer:

- `docs/migration/conventions.md` (379 lines) — the single migration recipe: Tier 1/2/3 target
  layouts walked through the three goldens (stripe declarative+writes, searxng read-only/no-auth,
  postgres Tier-3 native split), authoring rules (naming, x-secret, schema-as-projection,
  sync-mode derivation, api_surface minimal-honest depth per DECISIONS.md #4, docs.md headings),
  a full engine-dialect reference read directly from `engine/{bundle,interpolate,paginate,read,
  write,hooks,schema}.go` (template refs/filters, when-grammar incl. absent-key-falsy, all 6
  pagination types + none, param_format, computed_fields, conditional headers, client_filtered,
  MaxPages, write body construction, delete semantics), fixture rules (2-page-required, RFC3339
  cursor-fixture convention with the WHY, secret-scan ban), a 10-row parity-deviation ledger
  distilled from `waveF-b15-ledger.md`/`waveF-b16-ledger.md`/`waveF-b17-ledger.md` plus a note on
  the two gaps the `waveF-repair-ledger.md` closed, an escape-hatch decision tree with the exact
  blocker taxonomy, and a self-verify command block + forbidden-files list + no-commit rule.
- `docs/migration/result.schema.json` — draft-07 schema for the per-agent structured result
  (`connectors[]` with name/status/files_changed/streams_before/streams_after/
  write_actions_added/escape_hatches/fixtures_added/conformance/blockers/notes, plus optional
  `parity_deviations[]` and the optional Pass-B fields `api_surface_endpoints_total`/
  `endpoints_implemented`/`endpoints_skipped[]`) per orchestration-plan.md's "Per-agent task spec".
- `docs/migration/review.schema.json` — draft-07 schema for the adversarial reviewer verdict
  ({connector, verdict, findings[], checked{}}) per orchestration-plan.md's "Verification
  pyramid" layer 3 checklist.

## Grounding

Read before writing: PLAN.md's D-20 task spec and Wave F entries; orchestration-plan.md's
"Per-agent task spec" and "Verification pyramid" sections; all three golden bundles in full
(every file under `defs/stripe/**`, `defs/searxng/**`, `defs/postgres/**` +
`native/postgres/*.go`); `waveF-b15-ledger.md`, `waveF-b16-ledger.md`, `waveF-b17-ledger.md`,
`waveF-repair-ledger.md` (the parity-deviation ledger is a direct distillation of these, not
invented); `engine/{bundle,interpolate,paginate,read,write,hooks,schema}.go` in full (the dialect
reference section is written from the actual code, not from the design doc's abridged examples);
`cmd/connectorgen/validate.go` and `conformance/static.go` (rule names, closed exclusion
vocabulary, required doc headings — cross-checked, both files agree); design doc §B.6, §B.7, §E.1,
§F; DECISIONS.md (api_surface depth decision #4); `docs/prompts/universal-programming-loop-prompts.md`
(the executor/reviewer templates these two schemas serve).

## Verification

```
$ python3 -m json.tool docs/migration/result.schema.json >/dev/null && echo OK
OK
$ python3 -m json.tool docs/migration/review.schema.json >/dev/null && echo OK
OK
$ wc -l docs/migration/conventions.md
379 docs/migration/conventions.md
```

Every concrete file path cited in `conventions.md` was spot-checked to exist via `ls`/test -e
(all 27 checked paths present; the one forward-reference, `docs/migration/quarantine.json`, is
phrased as a future-generated artifact — matching its own listing in orchestration-plan.md's
"Artifacts" section as a wave-N-generated file, not claimed to exist today).

## Path guard

Touched only: `docs/migration/conventions.md`, `docs/migration/result.schema.json`,
`docs/migration/review.schema.json` (all new), and this ledger file. No other file read-modified;
no git commit made (orchestrator's responsibility per the no-commit rule this same doc states).

## Blockers

None.
