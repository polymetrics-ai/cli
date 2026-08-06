# Discussion log — provider-cited rate-limit declarations R1

The worker is autonomous and no human decision is needed for this bounded rollout. The delivery
brief resolves the material design choices:

- Provider sweep artifacts identify candidate connectors but are not rate-limit evidence.
- A value is declared only from a current provider-controlled, citable policy source.
- `unknown` is the required fallback whenever an exact policy or compatible scope cannot be
  established.
- The first release must include the optional `defs.FS` wildcard so declarations are shipped.
- The legacy stream page-loop throttle remains separate and untouched.
