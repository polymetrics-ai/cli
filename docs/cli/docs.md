```
NAME
  pm docs - generate CLI documentation

SYNOPSIS
  pm docs generate --dir <path>
  pm docs validate [--connectors-dir <path>]

DESCRIPTION
  Writes embedded command documentation as markdown files. Generation also
  writes connector manuals and connector SKILL.md files under a connector docs
  directory. By default, when --dir is docs/cli, connector docs are written to
  docs/connectors.

  Validation checks every registered connector has a generated MANUAL.md and
  SKILL.md with required human and agent sections. This is intended for CI hooks
  and local preflight checks before adding or changing connectors.

OPTIONS
  --dir path             command docs output directory
  --connectors-dir path  connector docs output directory

SECURITY
  Generated docs contain no credentials.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
