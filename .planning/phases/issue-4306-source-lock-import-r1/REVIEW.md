# Code review — source-lock operation import

## Scope

- `cmd/connectorgen/sourceimport.go`, command registration, source-import tests and synthetic fixtures
- migration adoption documentation and issue #4306 GSD/TDD evidence

## Inline review repair

The post-review findings were legitimate and repaired in the shared importer rather than at individual call sites.

- Trust boundary: source artifact URLs reuse the public-artifact URL and DNS/dial guard, reject userinfo/query/private destinations, and retain the no-redirect policy.
- Parsing and ownership: YAML now permits exactly one document; fixed paths reject whitespace and network-path forms; resolved definitions, bundle, sources, and lock paths each remain within their owner.
- Request contracts: effective operation parameters override path-item parameters, every path placeholder has one required path parameter, Swagger 2 JSON bodies inherit `consumes`, and objects require explicit fixed properties.
- Output and auth: only JSON media are marked JSON, non-text media are binary, ambiguous mixed media fail closed, and auth preserves OR/AND requirement groups including zero-scope schemes.
- Regression coverage: checked-in synthetic fixtures cover the added rejection classes and Swagger body projection; focused source-import tests exercise public-destination, redirect, and symlink containment guards.

## Automated review route

No PR exists in this Firstmate handoff lane, so GitHub-hosted Claude/Copilot review cannot yet run. The later PR owner must follow the repository automated-review routing contract; this inline review is not represented as GitHub review coverage.
