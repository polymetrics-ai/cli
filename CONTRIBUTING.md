# Contributing

Thanks for helping improve Polymetrics CLI. The fastest useful contribution is a focused bug fix, connector improvement, or documentation update with tests where behavior changes.

## Development Setup

Prerequisites:

- Go 1.25.11 or newer
- Node.js 22 and pnpm 11 for the website
- A POSIX shell for the smoke test

Clone and verify:

```bash
git clone https://github.com/polymetrics-ai/cli
cd cli
make verify
```

Website checks:

```bash
cd website
pnpm install --frozen-lockfile
pnpm run gen:website-data
pnpm run typecheck
pnpm run test:unit
pnpm run test:e2e
pnpm run build
```

## Pull Requests

- Keep pull requests narrowly scoped.
- Create branches as `<type>/<description>`, for example `feat/github-connector`, `fix/stripe-pagination`, or `docs/install-binaries`.
- Title PRs with [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/), for example `feat(connector): add linear` or `fix(github): repair pagination`.
- Add or update tests for behavior changes.
- Run `make verify` before requesting review.
- Do not include credentials, API tokens, private URLs, customer data, or generated local state.
- Update docs when CLI behavior, flags, output, connector setup, or supported workflows change.

## Release Versioning

Releases are generated from Conventional Commits after changes land on `main`.

- `fix(<connector>): ...` creates a patch release for connector updates, bug fixes, pagination changes, auth fixes, docs that ship with the binary, and other compatible repairs.
- `feat(connector): add <name>` creates a minor release for a new connector or new user-facing capability.
- Add `!` before the colon, or a `BREAKING CHANGE:` footer, for breaking changes that require a major release.
- Use other Conventional Commit types such as `docs:`, `ci:`, `test:`, `refactor:`, and `chore:` when the change should be categorized without implying a feature or bug-fix release.

## Connector Contributions

Connector work should follow the existing package structure in `internal/connectors`. Prefer copying a similar connector and adapting the auth, streams, pagination, schema, and tests.

Run:

```bash
go test ./internal/connectors/...
make verify
```

## Licensing

By contributing, you agree that your contribution is provided under the repository license in `LICENSE`.
