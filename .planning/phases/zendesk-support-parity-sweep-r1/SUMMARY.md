---
phase: zendesk-support-parity-sweep-r1
program: cli-top50-fixed-schema-sweep-r1
connector: zendesk-support
issue: 156
branch: fm/cli-top50-sweep-resume2-r1
coverage:
  documented_operations: 625
  carried_non_oas_rows: 6
  api_surface_rows: 631
  covered: 509
  blocked: 122
  commands_before: 631
  commands_after: 633
  commands_implemented: 475
  commands_partial: 36
  commands_planned: 122
  write_actions_before: 89
  write_actions_after: 173
  streams: 33
  reachability_verified: 511
  reachability_failures: 0
---

# zendesk-support — documented-operation parity

**625 documented operations, 509 reachable, 122 blocked with a named dependency each.**
Landing order #3 under the captain's largest-first reversal, behind github (1220) and
workday-rest (907).

## What this connector needed, and how it differed from the two before it

github and workday-rest were **enumeration** problems: the bundle did not know what the provider
documented. zendesk-support was the opposite. #3532 already shipped a complete, correctly counted
631-row inventory. What it shipped alongside was **509 rows disposed by one sentence** — "blocked by
default until shared foundation #2985" — and 509 `availability: planned` command placeholders.

So the inventory was truthful and the surface was useless. `AGENTS.md` says
`availability: implemented` is a claim the runtime has to honour; the inverse failure is a
permanently-`planned` surface that claims nothing and does nothing. This phase was **promotion**,
not enumeration, and the count never moved.

| | before | after |
| --- | ---: | ---: |
| api_surface rows | 631 | 631 |
| covered | 122 | **509** |
| blocked | 509 | **122** |
| blocked naming a real dependency | 0 | **122** |
| commands implemented | 95 | **475** |
| commands partial | 27 | 36 |
| commands planned | 509 | **122** |
| writes.json actions | 89 | **173** |
| verified reachable by running the binary | — | **511 / 511** |

## The count: 625, and the first ledger in this sweep that reconciles

Re-fetched at **1,701,930 bytes**, byte-identical to `MASTER-PLAN.json`, and re-derived rather than
carried forward: 434 paths, 625 operations, GET 325 / POST 111 / PUT 89 / DELETE 86 / PATCH 14.
Zero delta against the provider ledger — the first time in this sweep, after six wrong ones. It
reconciles because Zendesk publishes a machine-readable spec keyed one operation per (method, path)
rather than documentation pages. **That is evidence about this artifact, not grounds to trust the
next ledger.**

Finding 34 was applied rather than restated: the derivation is run through the red test's own rules
(`?`, `*`, space, duplicate pairs) and exits non-zero before its number can be adopted. Neither
collapse that shrank workday-rest exists here.

## The four judgements

**Read vs write.** METHOD decides, with eight pinned read-shaped POSTs. Two traps, both real:
three POSTs match a naive read keyword through the *resource* name (`saved_searches`,
`task_list_templates`, `task_lists`) while literally saying "Creates a …" — writes. And the shipped
bundle declared two async exports `kind: rest_read`; Zendesk's own text says they "enqueue a job"
and "send the requester an email containing a link to the CSV file", and `/audit_logs/export`
documents only a 202 Accepted with no body. Enqueuing a job and sending mail are side effects. They
are writes, matching how this sweep classified GitHub's analogous exports in the same batch.

**Stream vs direct read.** The 33 shipped fixture-backed streams keep their rows (greenhouse
finding 21). Every other read is a bounded direct read.

**Binary detection.** Exactly one `application/octet-stream` operation. The trap here runs
*backwards* to workday-rest's: `PUT /api/v2/brands/{brand_id}` declares `image/jpg` and `image/png`
because updating a brand echoes its logo. A "declares a non-JSON media type" rule alone would ship
it as a download and silently drop the mutation. Binary is GET-only, which is what makes the trap
checkable instead of a matter of taste. workday-rest's trap was paths that *sounded* binary and
declared JSON; together the pair is the rule — read the media types, and require GET.

**Named-dependency blocking.** All 122, by cause:

| class | n | the component that refuses it |
| --- | ---: | --- |
| no request contract | 106 | the OAS declares no `requestBody`; deriving one means inventing a payload |
| empty record contract | 10 | `engine.PreflightWriteAction` refuses a schema admitting only `{}` |
| file upload | 5 | no bounded multipart file-transfer executor exists |
| credential body | 1 | no stdin/env-sourced secret input for a request-body credential |

The no-body claim was **verified against the artifact**, not inherited: of the writes the bundle
said had no body, the spec agreed on all but one — the multipart brand logo, which is a file upload.

## Three things worth carrying to the next connector

**1. `covered_by.writes` earned its keep.** `PUT /api/v2/tickets/update_many` and
`PUT /api/v2/users/update_many` have genuinely distinct union arms — one applies a single set of
values to many ids, the other carries per-record values. Each became its own named action on one
documented path through the plural array github landed in this sweep. This is the first connector to
use it for its stated purpose, and it confirms finding 31 generalised rather than solving github's
problem. Zendesk writes these unions as a bare required-list per arm over **parent-level**
properties, so an arm read in isolation requires a field it never defines; each arm is rebuilt as
the parent shape narrowed to its own field.

**2. A union is not always a union.** Both filtered-search bodies are the *same* open object twice,
labelled Basic and Complex — documentation prose. The nested `oneOf`s on `brand_id` and `filter` are
type variants, which draft-07 expresses as a type array, keeping every documented arm reachable. An
arm with its own `required` list is a different contract and the generator refuses to flatten it.
The three cases are distinguishable mechanically, and conflating them either ships a lie or blocks
an operation that was always reachable.

**3. Test the built artifact, not its inputs.** The empty-record-contract guard was written first
against the inputs — has a body? has a path variable? — and passed three operations that declare a
body schema containing no properties. Runtime preflight caught them. The guard now tests the schema
it actually built, which is the only thing the runtime ever sees. The same shape of error appeared
in finding 34 one connector earlier: a derivation that was never run through the rules its own test
enforces.

## Known-unmet, carried deliberately

`TestGoldenTranscripts` fails in 11 subtests and has since before github. It is discharged by the
**end-of-sweep** artifact regeneration, never per connector — finding F6: a per-connector
`pm docs generate` run rewrites ~1,031 files of pre-existing `main` drift. See `VERIFICATION.md`
for the measurement proving this phase adds no new failure to that set.
