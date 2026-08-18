# Census — 277 supplied GitHub product defects

## Result

| Primary root-cause class | Commands | Share |
| --- | ---: | ---: |
| `integer_id_scientific_notation` | 147 | 53.1% |
| `missing_request_payload` | 56 | 20.2% |
| `wrong_api_path` | 9 | 3.2% |
| `false_success` | 11 | 4.0% |
| `other` | 54 | 19.5% |
| **Total** | **277** | **100.0%** |

The measured census does **not** support extrapolating the original 77% sample
to roughly 200 commands. Precision loss is still the largest class, but the
complete supplied inventory identifies 147 primary instances (53.1%).

## Method and precedence

The source inventory is
`product-defect-commands.json` SHA-256
`6bd82263d693f378c6b88263d84d288108f25992c93d543d9bdb77681ac609d8`.
It contains 88 slice-2, 39 slice-3, 69 slice-4, and 81 slice-5 commands.
The slice manifests were joined by exact command path; the committed slice
ledgers supplied reasoned ordinals where the inventory's reason was blank.

Each command has exactly one primary class. A machine-readable live reason has
first precedence, then a reasoned slice-ledger class, then a structural check of
the current command/write declarations. For unreasoned rows, an integer flag
feeding a write-action path is precision loss; a provider mutation with no
non-path body field is missing payload; an Actions secret/variable command on
`/agents/` is wrong path; and a DELETE whose declared missing response is counted
as written is false success. Rows without one of those independently inspectable
signatures remain `other`. An overlapping secondary defect does not double-count
a command; for example, an integer-ID command on a wrong endpoint remains in its
recorded live class.

The source hashes used to map blank-reason slice ordinals were:

- slice 2: `b01277378c4b8103cb3d6f2d747728345b77787982ad308065672d4f7f7810e8`
- slice 4: `e8da76b3973c452275732ce027be36d0f5f39a1f7cbea37b13cb45a54dcccadb`
- slice 5: `4e4fe04a23f8fd787e36584bba1ae304c3c341262aae04de5fe1ca1835249662`

## Command-level assignments

### `integer_id_scientific_notation` (147)

```text
actions add-repo-access-to-self-hosted-runner-group-in-org
actions add-selected-repo-to-org-secret
actions add-selected-repo-to-org-variable
actions add-self-hosted-runner-to-group-for-org
actions approve create
actions artifacts delete
actions delete-custom-image-version-from-org
actions delete-hosted-runner-for-org
actions delete-self-hosted-runner-group-from-org
actions deployment_protection_rule create
actions disable set
actions disable-selected-repository-github-actions-organization
actions disable-selected-repository-self-hosted-runners-organization
actions enable set
actions enable-selected-repository-github-actions-organization
actions enable-selected-repository-self-hosted-runners-organization
actions force-cancel create
actions labels delete
actions labels delete-2
actions labels set
actions logs delete
actions pending_deployments create
actions remove-all-custom-labels-from-self-hosted-runner-for-org
actions remove-custom-label-from-self-hosted-runner-for-org
actions remove-repo-access-to-self-hosted-runner-group-in-org
actions remove-selected-repo-from-org-secret
actions remove-selected-repo-from-org-variable
actions remove-self-hosted-runner-from-group-for-org
actions rerun create
actions rerun-failed-jobs create
actions update-hosted-runner-for-org
activity delete-thread-subscription
activity mark-thread-as-done
activity mark-thread-as-read
activity set-thread-subscription
agents add-selected-repo-to-org-secret
agents add-selected-repo-to-org-variable
agents remove-selected-repo-from-org-secret
agents remove-selected-repo-from-org-variable
app access_tokens create
apps add-repo-to-installation-for-authenticated-user
apps remove-repo-from-installation-for-authenticated-user
apps suspend-installation
apps unsuspend-installation
autolinks delete
campaigns delete-campaign
campaigns update-campaign
check-runs rerequest create
check-runs update
check-suites rerequest create
code-security attach-configuration
codespaces add-selected-repo-to-org-secret
codespaces remove-selected-repo-from-org-secret
comments reactions create
comments reactions delete
copilot disable-coding-agent-for-selected-repository-for-organization
copilot enable-coding-agent-for-selected-repository-for-organization
copilot-spaces create-resource-for-user
copilot-spaces delete-for-org
copilot-spaces delete-for-user
copilot-spaces delete-resource-for-org
copilot-spaces delete-resource-for-user
copilot-spaces remove-collaborator-for-org
copilot-spaces remove-collaborator-for-user
copilot-spaces update-collaborator-for-org
copilot-spaces update-collaborator-for-user
copilot-spaces update-for-org
copilot-spaces update-for-user
copilot-spaces update-resource-for-org
copilot-spaces update-resource-for-user
dependabot add-selected-repo-to-org-secret
dependabot remove-selected-repo-from-org-secret
deployments statuses create
environments deployment-branch-policies delete
environments deployment-branch-policies set
gists delete-comment
gists update-comment
import authors update
invitations delete
invitations update
issue close
issue lock
issue reopen
issue unlock
issues approve-suggestion
issues blocked_by delete
issues dismiss-suggestion
issues pin set
issues reactions create
issues reactions delete
issues reactions delete-2
migrations delete-archive-for-authenticated-user
migrations delete-archive-for-org
migrations unlock-repo-for-authenticated-user
migrations unlock-repo-for-org
orgs disable-selected-repository-immutable-releases-organization
orgs enable-selected-repository-immutable-releases-organization
orgs redeliver-webhook-delivery
orgs update-issue-field
orgs update-issue-type
orgs update-webhook
orgs update-webhook-config-for-org
pr comment
pr lock
pr merge
pr reopen
pr unlock
pr update-branch
projects delete-item-for-org
projects delete-item-for-user
projects update-item-for-org
projects update-item-for-user
pulls merge-async
pulls reactions create
pulls reactions delete
pulls replies create
pulls reviews delete
pulls reviews set
release delete
release delete-asset
release edit
releases assets view-3
repo ruleset delete
repo ruleset update
repos accept-invitation-for-authenticated-user
repos delete-org-ruleset
repos update-org-ruleset
security-advisories forks create
teams add-member-legacy
teams add-or-update-membership-for-user-legacy
teams add-or-update-repo-permissions-legacy
teams delete-legacy
teams remove-member-legacy
teams remove-membership-for-user-legacy
teams remove-repo-legacy
teams update-legacy
users delete-attestations-by-id
users delete-gpg-key-for-authenticated-user
users delete-public-ssh-key-for-authenticated-user
users delete-ssh-signing-key-for-authenticated-user
users projects fields create-iteration
users projects fields create-new-field
users projects fields create-single-select
users projects items create-by-id
users projects items create-by-repo-number
webhook create-2
webhook create-3
```

### `missing_request_payload` (56)

```text
actions create-remove-token-for-org
actions generate-jitconfig create
actions permissions set
actions permissions set-2
actions permissions set-3
actions permissions set-4
actions permissions set-6
actions registration-token create
actions remove-token create
actions retention-limit set
autolinks create
branches apps create
branches apps set
branches contexts create
branches contexts set
branches enforce_admins create
branches protection set
branches rename create
branches required_pull_request_reviews update
branches required_signatures create
branches required_status_checks update
branches teams create
branches teams set
branches users create
branches users set
check-runs create
check-suites preferences update
codespaces create
dependency-graph snapshots create
environments deployment-branch-policies create
git blobs create
git commits create
interactions set-restrictions-for-org
issues priority update
issues reactions create-2
issues sub_issue delete
issues sub_issues create
merge-upstream create
migrations start-for-authenticated-user
migrations start-for-org
properties values update
releases generate-notes view
repo update
secret set-3
secret set-5
security-advisories create
security-advisories reports create
teams create
transfer create
users block
users follow
variable create
variable create-2
variable update
variable update-2
webhook update
```

### `wrong_api_path` (9)

```text
agents set-selected-repos-for-org-secret
agents set-selected-repos-for-org-variable
agents update-org-variable
projects create-draft-item-for-authenticated-user
secret delete-6
secret set-6
variable create-3
variable delete-4
variable update-3
```

### `false_success` (11)

```text
actions delete-org-secret
issue transfer
orgs create-webhook
orgs delete-webhook
teams delete-in-org
teams remove-membership-for-user-in-org
teams remove-repo-in-org
users delete-attestations-by-subject-digest
users delete-social-account-for-authenticated-user
users unblock
users unfollow
```

### `other` (54)

```text
actions caches delete
actions create-hosted-runner-for-org
actions create-self-hosted-runner-group-for-org
actions generate-runner-jitconfig-for-org
actions set-actions-cache-retention-limit-for-enterprise
actions set-artifact-and-log-retention-settings-organization
actions set-fork-pr-contributor-approval-permissions-organization
actions set-selected-repositories-self-hosted-runners-organization
actions update-org-variable
agent-tasks create-task-in-repo
branches apps delete
branches contexts delete
branches enforce_admins delete
branches protection delete
branches required_pull_request_reviews delete
branches required_signatures delete
branches required_status_checks delete
branches restrictions delete
branches teams delete
branches users delete
codespaces set-codespaces-access
copilot-spaces create-for-org
dependabot create-or-update-org-secret
dependabot set-repository-access-default-level
dependabot set-repository-access-default-level-for-enterprise
enterprise-teams create
git tags create
import delete
interactions update-pull-request-creation-cap-for-org
orgs enable-or-disable-security-product-on-all-org-repos
orgs ping-webhook
pr create
pr revert
private-registries create-org-private-registry
pull-request-stacks create
repo delete
repo delete-2
repos create-for-authenticated-user
secret delete
secret delete-2
secret delete-3
secret delete-5
secret set
secret set-2
subscription delete
teams add-or-update-membership-for-user-in-org
teams add-or-update-repo-permissions-in-org
teams update-in-org
user emails add
users attestations delete-by-attestation-ids
users attestations delete-by-subject-digests
users create-gpg-key-for-authenticated-user
users set-primary-email-visibility-for-authenticated-user
users update-authenticated
```

## Scope statement

This delivery fixes the four named classes. The 54 `other` rows are deliberately
not claimed fixed: their measured reasons include provider deprecation, success
responses the CLI rejects, unprovable ephemeral outputs, typed string-ID drift,
anonymous-only endpoints, and command-specific body/GraphQL semantics that are
not one of the four authorized root causes. They remain declared and will retain
their exact certification result until separately repaired.
