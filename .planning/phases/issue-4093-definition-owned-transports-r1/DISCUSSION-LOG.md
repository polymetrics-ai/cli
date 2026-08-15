# Discussion log — #4093

`discuss-phase --auto` was resolved through the repository adapter and executed
inline. The issue and firstmate brief settle all material choices:

- the schema is optional and versioned, not a new required bundle file;
- failures are fail-closed and atomic with respect to executor registration;
- definitions select closed executors and externally admitted evidence, while
  provider-specific construction stays in named factories;
- GitHub and PostgreSQL declarations reside only in their own `defs/` bundles;
- `change_capture` cannot be a destination transport mode;
- use the supplied live PostgreSQL Docker endpoint for the native proof.

No product decision remains open.
