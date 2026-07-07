# Verification

## Planned Commands

```bash
git diff --check
find .agents/agentic-delivery .planning/phases/cross-connector-rollout -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
jq empty .planning/phases/cross-connector-rollout/RUN-STATE.json
```

## Results

- PASS: `git diff --check -- .agents/agentic-delivery .agents/skills/caveman .planning/phases/cross-connector-rollout`
- PASS:
  `find .agents/agentic-delivery .planning/phases/cross-connector-rollout -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'`
- PASS: `jq empty .planning/phases/cross-connector-rollout/RUN-STATE.json`

## Gate Notes

- Full Go verification is not required for this docs-only slice because no Go, connector, or build
  files are edited.
- CodeRabbit manual review must not be triggered for this stacked sub-PR.
