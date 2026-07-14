```
NAME
  pm docs - generate CLI documentation

SYNOPSIS
  pm docs generate --dir <path>
  pm docs validate [--connectors-dir <path>]
  pm docs generate --dir <path> --connectors-dir <path> --connector <name> [--cli-connector <name>]

DESCRIPTION
  Writes embedded command documentation as markdown files. Generation also
  writes connector manuals under a connector docs directory. By default, when
  --dir is docs/cli, connector docs are written to docs/connectors.

  Validation checks every registered connector has a generated MANUAL.md with
  required human and agent workflow sections. This is intended for CI hooks and
  local preflight checks before adding or changing connectors.

  Repeat --connector to generate or exactly validate only those connector
  manuals, skills, and CLI pages. Repeat --cli-connector for additional CLI
  pages without rewriting those connectors' manuals or skills.

OPTIONS
  --dir path             command docs output directory
  --connectors-dir path  connector docs output directory
  --connector name       select a connector manual, skill, and CLI page
  --cli-connector name   select only an additional connector CLI page

SECURITY
  Generated docs contain no credentials.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
