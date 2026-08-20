# Code review — source-lock operation import

## Scope

- `cmd/connectorgen/sourceimport.go`, command registration, source-import tests and synthetic fixtures
- migration adoption documentation and issue #4306 GSD/TDD evidence

## Inline review repair

The post-review findings were legitimate and repaired in the shared importer rather than at individual call sites.

- Trust boundary: source artifact URLs reuse the public-artifact URL and DNS/dial guard, reject userinfo/query/private destinations, and retain the no-redirect policy.
- Parsing and ownership: YAML now permits exactly one document; fixed paths reject whitespace and network-path forms; resolved definitions, bundle, sources, and lock paths each remain within their owner.
- Request contracts: effective operation parameters override path-item parameters, every path placeholder has one required path parameter, Swagger 2 JSON bodies inherit `consumes`, and objects require explicit fixed properties.
- Output and auth: only JSON media are marked JSON, non-text media are binary, per-status success/error media remain independent, and auth preserves OR/AND requirement groups including zero-scope schemes.
- Parser and references: duplicate JSON/YAML object members fail with escaped source pointers; references resolve only in OpenAPI grammar positions, preserving literal provider `$ref` fields and valid 3.1 siblings.
- Deferred metadata: webhooks, callbacks, route-server layers, path extensions, and unsupported parameter serializations are retained as canonical merge-blocked foundation gaps rather than silently removed.
- Bounds and amplification: dynamic-reference vocabularies fail with named provenance, semantic finite bounds reject null/non-finite or contradictory limits, and an aggregate compact descriptor budget blocks reference amplification before retention.
- Regression coverage: checked-in synthetic fixtures and focused source-import tests cover the added ambiguity, metadata, mixed-media, and amplification contracts alongside public-destination, redirect, and symlink containment guards.
- Grammar completeness: Link and Example references resolve only in their defined positions and type-check their targets; content parameters have exactly one validated representation; `prefixItems` and unsupported JSON Schema applicators fail closed.
- Form and identity integrity: exact OpenAPI/Swagger versions are retained and validated, Swagger host/basePath/root-and-operation schemes remain merge-blocked route evidence, source identities are connector-global across outbound and inbound descriptors, and YAML keys must remain strings before normalization.
- Expansion safety: schema children reserve deterministic object/document capacity before cloning or retention, with typed limit errors for reference amplification.
- Semantic closure: finite composite enums are recognized as bounded contracts without permitting dynamic children, while OpenAPI-only reference positions remain form-scoped and Encoding Header references are resolved before retention.

## Automated review route

No PR exists in this Firstmate handoff lane, so GitHub-hosted Claude/Copilot review cannot yet run. The later PR owner must follow the repository automated-review routing contract; this inline review is not represented as GitHub review coverage.
