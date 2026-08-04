# Test plan

| Behavior | First red evidence | Green evidence |
| --- | --- | --- |
| `media.download` is command-reachable through `binary_download` | Focused commandrunner/CLI test or preflight failure before metadata is promoted | Targeted test plus `TestEveryImplementedCommandPassesRuntimePreflight` |
| `reports.query` remains honestly blocked | Surface inspection verifies it has no executable `provider_search`/`rest_write` mapping | Surface validator and command help do not claim reachability |
| Declared request fields have evidence | Citation matrix review detects every missing field | Connector validation and matrix accounting show all fields cited or tier-5 deferred |
| Generated surface stays in parity | `surface-sync --check` fails before regeneration when input drift exists | `surface-sync`, `--check`, and website data generation succeed |

No credentialed provider execution or reverse-ETL execution is in scope.
