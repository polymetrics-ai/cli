Refs #4015

## Slice 0: generated mutation inventory only

This extends the landed `connectorgen certification-candidates` projection; it
does not add a parallel projection and it does not invoke a credential,
provider endpoint, fixture, reverse plan, preview, approval, execution, or
evidence publication.

`internal/connectors/defs/github/certification.json` now declares the mutation
cohort and containment families. Shared Go derives each candidate's command
tokens, declaration identity, executor, required flags, credential flag, JSON
mode, typed input slots, address, and fixture provenance from `cli_surface.json`
plus the referenced operation or write action. No connector identifier or
connector allowlist was added to shared Go.

## Measured accounting

Measured with:

```sh
jq '{candidate_count: (.mutation_candidates | length), generated: ([.mutation_candidates[] | select(.generated == true)] | length), manual: ([.mutation_candidates[] | select(.generated != true)] | length), by_intent: (reduce .mutation_candidates[] as $candidate ({}; .[$candidate.intent] = (.[$candidate.intent] // 0) + 1)), by_fixture: (reduce .mutation_candidates[] as $candidate ({}; .[$candidate.fixture.strategy] = (.[$candidate.fixture.strategy] // 0) + 1)), by_classification: (reduce .mutation_candidates[] as $candidate ({}; .[$candidate.classification.code] = (.[$candidate.classification.code] // 0) + 1))}' internal/connectors/defs/github/certification.json
go run ./cmd/connectorgen certification-sweep --connector github --check
```

The full-surface CRUD calculation was rerun from the current sweep ledger with:

```sh
jq -c '
  def collection: split("/") | map(select(length > 0)) |
    if (length > 0 and (.[-1] | test("^\\{[^}]+\\}$"))) then .[0:-1] else . end |
    "/" + join("/");
  [.commands[] | select(.status == "fixture_required" and (.api_surface | length) > 0) |
   . as $command | .api_surface[] |
   {command: $command.path, method: (.method | ascii_upcase), collection: (.path | collection)}] as $addresses |
  ($addresses | map(select(.method == "POST" or .method == "PUT") | .collection) | unique) as $provisioned |
  {fixture_required_with_api_surface: ($addresses | map(.command) | unique | length),
   self_provisioning_exact_collection: ([$addresses[] | select(.collection as $collection | $provisioned | index($collection))] | map(.command) | unique | length),
   named_exceptions: (($addresses | map(.command) | unique | length) - ([$addresses[] | select(.collection as $collection | $provisioned | index($collection))] | map(.command) | unique | length))}
' internal/connectors/defs/github/certification-sweep.json
```

| Account | Count |
| --- | ---: |
| `direct_write` candidates | 279 |
| `reverse_etl` candidates | 577 |
| Mutation candidates | **856** |
| Generated / manual candidates | **856 / 0** |
| Derived REST collection cycles / named provisioning exceptions | **489 / 367** |
| `contained` | 793 |
| `real_money` | 15 |
| `real_people` | 38 |
| `public_visibility` | 10 |
| `third_party_scope` | 0 |
| `unassessed` | 0 |
| Classification buckets | **856** |
| Entire declared command surface | **1,571** |

Arithmetic: `279 + 577 = 856`; `793 + 15 + 38 + 10 + 0 + 0 = 856`.
The canonical certification sweep reports the complete 1,571-command surface.

There are no manual candidates. The generator nevertheless permits an override
only when it names an exact generated command and provides `override_reason`;
an invented or duplicate manual command is rejected.

## Derived CRUD fixture provenance

The candidate projection does not point at the existing long-lived write-wave
fixture. For a REST candidate with a `cli_surface` address, it derives its
collection by stripping only a terminal `{id}`-style segment. Every POST or PUT
candidate on that resulting collection becomes a listed provisioner for its
siblings, and derived collections are ordered by URL depth so parents precede
children.

The current GitHub sweep independently measures 1,355 fixture-required commands
with an API surface: 969 have a POST or PUT on this exact derived collection,
leaving 386 named exceptions in the full surface. In the 856-mutation Slice 0
cohort, 489 REST candidates derive a collection cycle. The remaining 367 are
explicitly labelled `named_exception`: shared GraphQL transport is not treated
as one giant resource collection, a friendly alias without `api_surface` cannot
borrow a raw action path, and a REST collection without POST/PUT records the
specific absence. Their stable codes are `graphql_transport_not_collection`
(279), `collection_without_creator` (74), and `missing_api_surface` (14).
None falls back to a static fixture.

This remains an inventory only. Slice 0 makes no provider contact, so it neither
executes nor certifies a mutation. In a later authorized live slice, executing
with produced-value assertions and verified cleanup must publish the matching
certification in the same act; withholding because another mechanism owns the
record would be a blocker, not a staging state.

## Containment classification

Classification is about whether an effect can escape the disposable boundary,
not whether a verb sounds reversible. The connector-owned catalog explicitly
identifies paid-seat/sponsorship changes, people/invitation changes, and public
publication changes. Everything reported as `contained` retains evidence that a
future live slice must bind its target to the disposable certification identity
or organisation and destroy its run-owned container afterwards. Irreversible
actions are not excluded solely for being irreversible.

`third_party_scope` is a supported, separate machine code even though the
current declared candidate cohort contains no member of that class. It is not a
captain-decidable live target: third-party repositories and organisations remain
forbidden.

## Classifier refusal demonstration

`TestBuildGeneratedMutationCandidatesClassifiesEscapeAndFailsClosed` constructs
only in-memory declarations. It passes only when:

| Constructed declaration | Required result |
| --- | --- |
| disposable widget mutation | `contained` |
| paid-seat mutation | `real_money` |
| outside invitation | `real_people` |
| public publication | `public_visibility` |
| third-party repository target | `third_party_scope` |
| unmatched declaration | `unassessed` |

The final row is the honest-accounting guard: no family match can default to
`contained`. The test also rejects a family selector that names a command,
operation, or write action outside the declared cohort, so a typo cannot fall
through into a broad contained family.

## Rules for the rulebook

- Certify containment, not reversibility. An irreversible mutation is eligible
  only if its entire effect remains in a fresh disposable container and cleanup
  destroys that container.
- Treat the four escape codes as explicit policy boundaries: money, people,
  public visibility, and third-party scope. A third-party target is forbidden,
  not a future approval path.
- Keep cohort membership and containment evidence in connector definitions;
  generic shared code must not name a connector or carry connector allowlists.
- Project command mechanics from declarations. Manual candidates are bounded,
  exact-command exceptions with a written reason, never a broad escape hatch.
- Derive the REST CRUD fixture cycle from the endpoint tree. A sibling POST or
  PUT provisions its collection; strip only a terminal item parameter, order
  collections by URL depth, and make every non-derivable case a named exception.
- Never mistake GraphQL's shared transport endpoint for a resource collection;
  require a separately derived semantic cycle before later live work.
- Fail closed. A missing, unmatched, duplicated, or non-cohort classification
  must be `unassessed` or an error, never an implicit contained verdict.
- If a later slice executes and verifies a candidate, certify it in that same
  act. Withhold only when it did not execute, did not assert produced values,
  did not verify cleanup, or cannot state a verified claim.

## Verification

Passed:

- focused mutation projection, classifier-refusal, fail-closed, and CRUD-cycle
  tests in `./cmd/connectorgen`;
- isolated `TestCertificationMatrixRejectsDatabaseWriteStubs` in 979.333s;
- candidate and sweep generation twice, then both `--check` commands;
- `pnpm --dir website run gen:docs`, `pnpm --dir website run
  gen:website-data`, and `go run ./cmd/pm skills generate --dir docs/skills`,
  each twice with an empty combined diff on the second run;
- `make verify` end-to-end, including full Go tests, build, docs, smoke, lint,
  validation, candidate/sweep freshness, and a clean whole-tree boundary scan.

The first aggregate attempt timed out under concurrent full-suite load in
unrelated matrix/app/boundary/conformance packages. No test was changed or
skipped; the exact matrix test then passed alone and the complete cached retry
passed. Full command/result detail is in the committed GSD
`VERIFICATION.md`. All projection/generator commands are local declaration
processing only; no credential was needed or read.
