# Self-hosted CI runner policy

## Blocking GitHub configuration

`.github/workflows/runner-selection.yml` is routing policy, not a fork-security
boundary: a fork pull request can modify its workflow source before a job is
scheduled. CI routing is unsafe until every control below has been applied and
verified by a GitHub organization owner.

1. In **Organization settings → Actions → General → Fork pull request
   workflows**, select **Require approval for all external contributors**. Do
   not use either first-time-contributor option, because a previously merged
   external contributor would otherwise run without a new approval.
2. In **Organization settings → Actions → Runner groups**, move the live CLI
   runners into a dedicated non-default `polymetrics-cli` runner group. Set
   **Repository access** to **Selected repositories** and select only
   `polymetrics-ai/cli`; do not leave those runners in a default or
   all-repositories group.
3. Before selecting **Approve workflows to run** for a fork pull request,
   inspect every change under `.github/workflows/`. Do not approve a run whose
   proposed workflow can request `self-hosted`, `polymetrics-cli`, or another
   protected runner label.

GitHub's [fork workflow approval guidance](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/approve-runs-from-forks)
and [runner group access documentation](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/manage-access)
describe the corresponding settings.

## Routing contract

After the blocking configuration is in place, the shared selector may route a
same-repository pull request from `karthik-sivadas` or
`alfred-polymetrics-ai` to `polymetrics-cli`. It routes fork pull requests and
all non-pull-request events, including website deployment, to GitHub-hosted
`ubuntu-latest`.
