# Licensing

License names in this document use [SPDX identifiers](https://spdx.org/licenses/).

## Default License

The default license is `AGPL-3.0-only`, provided in [LICENSE](LICENSE). It
applies to every file in this repository unless a more specific license is
provided in a nearer `LICENSE` file or an explicit file notice.

## Connector Definitions

The declarative connector definitions use a permissive license so contributors
and downstream tools can inspect, improve, and reuse the catalog independently
of the Polymetrics runtime.

| Path | SPDX license | License text |
| --- | --- | --- |
| `internal/connectors/defs/**` | `MIT` | [internal/connectors/defs/LICENSE](internal/connectors/defs/LICENSE) |

The MIT boundary applies only to files inside `internal/connectors/defs/`.
Connector engines, generators, hooks, native implementations, the CLI,
website, and all other repository files remain under the default license unless
another license notice says otherwise. Distributing connector definitions with
the core does not remove either license's requirements.

## Contributions

Contributions are licensed under the license governing their destination path:

- Contributions under `internal/connectors/defs/` are provided under `MIT`.
- Contributions elsewhere are provided under `AGPL-3.0-only` unless a more
  specific license notice applies.

By contributing, you represent that you have the right to provide the
contribution under the applicable license.

## Third-Party Material

Third-party and vendored material remains under its original license. A nested
license, attribution file, or file header takes precedence for that material.

## Commercial Use And Separate Terms

The AGPL permits commercial use subject to its conditions. This repository
does not grant a separate commercial license. Organizations that need different
terms may contact Polymetrics AI about a separate written agreement.

## Trademarks

The software licenses do not grant rights to use Polymetrics AI names, logos,
or trademarks except as required for reasonable and customary description of
the origin of the software.
