package github

import "polymetrics.ai/internal/connectors"

// allSyncModes mirrors connectors.allSyncModes (unexported there). Kept local so
// the github package does not depend on an unexported connectors helper.
func allSyncModes() []string {
	return []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	}
}

// readSourceSyncModes mirrors connectors.readSourceSyncModes (unexported there).
func readSourceSyncModes() []string {
	return []string{"full_refresh", "incremental"}
}

// Manifest implements connectors.ManifestProvider for the GitHub connector. The
// connectors registry and guide renderer detect this method via the
// ManifestProvider interface.
func (g Connector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: g.Metadata(),
		ConfigFields: []connectors.ConfigField{
			{Name: "repository", Description: "Repository in owner/repo format.", Required: true},
			{Name: "base_url", Description: "GitHub API base URL override for tests or GitHub Enterprise."},
			{Name: "auth_type", Description: "Authentication mode: auto, public, token, or github_app.", Default: "auto"},
			{Name: "app_id", Description: "GitHub App ID or client ID for auth_type=github_app."},
			{Name: "installation_id", Description: "GitHub App installation ID for auth_type=github_app."},
			{Name: "installation_repositories", Description: "Optional comma-separated repository names for a restricted installation token."},
			{Name: "installation_repository_ids", Description: "Optional comma-separated repository IDs for a restricted installation token."},
			{Name: "installation_permissions", Description: "Optional JSON object of requested GitHub App installation-token permissions."},
			{Name: "per_page", Description: "Records per page.", Default: "100"},
			{Name: "max_pages", Description: "Maximum pages; use 0, all, or unlimited to exhaust the stream.", Default: "1"},
			{Name: "state", Description: "Issue, pull request, or milestone state filter where the GitHub API supports it.", Default: "all"},
			{Name: "sort", Description: "Issue, pull request, or milestone sort field where supported."},
			{Name: "direction", Description: "Sort direction where supported."},
			{Name: "since", Description: "Lower-bound timestamp for issues, issue_comments, pull_request_review_comments, and commits."},
			{Name: "until", Description: "Upper-bound timestamp for commits."},
			{Name: "sha", Description: "Branch or commit SHA filter for commits."},
			{Name: "path", Description: "Repository path filter for commits."},
			{Name: "author", Description: "Author filter for commits."},
			{Name: "committer", Description: "Committer filter for commits."},
			{Name: "actor", Description: "Actor filter for workflow_runs."},
			{Name: "branch", Description: "Branch filter for workflow_runs."},
			{Name: "event", Description: "Event filter for workflow_runs."},
			{Name: "status", Description: "Status filter for workflow_runs."},
			{Name: "created", Description: "Created-at filter for workflow_runs."},
			{Name: "head_sha", Description: "Head SHA filter for workflow_runs."},
			{Name: "check_suite_id", Description: "Check suite ID filter for workflow_runs."},
		},
		SecretFields: []connectors.SecretField{
			{Name: "token", Description: "GitHub token. Can be classic PAT, fine-grained PAT, OAuth token, GitHub Actions GITHUB_TOKEN, or installation access token.", Required: false},
			{Name: "personalAccessToken", Description: "Alias for token.", Required: false},
			{Name: "oauthToken", Description: "Alias for token when supplied by an OAuth app.", Required: false},
			{Name: "accessToken", Description: "Alias for token.", Required: false},
			{Name: "installationToken", Description: "Alias for token when an installation token is generated outside pm.", Required: false},
			{Name: "githubToken", Description: "Alias for token when passing GitHub Actions GITHUB_TOKEN.", Required: false},
			{Name: "private_key", Description: "GitHub App PEM private key for auth_type=github_app.", Required: false},
			{Name: "private_key_base64", Description: "Base64-encoded GitHub App PEM private key for auth_type=github_app.", Required: false},
		},
		AuthModes:       githubAuthModeSpecs(),
		Streams:         githubStreams(),
		WriteActions:    githubWriteActionSpecs(),
		SyncModes:       allSyncModes(),
		SourceSyncModes: readSourceSyncModes(),
		Pagination: connectors.PaginationSpec{
			Type:           "page",
			PageSizeField:  "per_page",
			PageLimitField: "max_pages",
			DelayField:     "page_delay",
			DefaultLimit:   "1",
		},
		Risk: connectors.RiskSpec{
			Read:     "external API read",
			Write:    "external GitHub API mutation",
			Mutation: "creates or changes issues, pull requests, comments, reviewers, labels, milestones, releases, workflow runs, or repository contents",
			Approval: "reverse ETL plan approval required before writes",
		},
	}
}
