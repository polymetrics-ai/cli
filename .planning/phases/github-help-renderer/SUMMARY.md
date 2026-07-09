# Summary

Implemented the first #35 green slice:

- Added an optional connector `CommandSurfaceProvider` for docs/help metadata.
- Exposed parsed `cli_surface.json` from engine-backed connectors without changing runtime dispatch.
- Rendered GitHub command-surface help through the existing connector manual path.
- Added website generated `cliSurface` data and a connector-page command-surface section.
- Updated generated website connector data.

No runtime `pm github ...` command dispatch was added.
