package github

import (
	"net/url"
	"strings"

	"polymetrics/internal/connectors"
)

func githubStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "repository", Description: "Repository metadata.", PrimaryKey: []string{"node_id"}, CursorFields: []string{"updated_at"}, Fields: githubRepositoryFields()},
		{Name: "issues", Description: "Repository issues. Pull requests returned by the GitHub issues endpoint are filtered out.", PrimaryKey: []string{"node_id"}, CursorFields: []string{"updated_at"}, Fields: githubIssueFields()},
		{Name: "pull_requests", Description: "Repository pull requests.", PrimaryKey: []string{"node_id"}, CursorFields: []string{"updated_at"}, Fields: githubPullRequestFields()},
		{Name: "branches", Description: "Repository branches.", PrimaryKey: []string{"name"}, Fields: githubBranchFields()},
		{Name: "commits", Description: "Repository commits.", PrimaryKey: []string{"sha"}, CursorFields: []string{"commit_committer_date"}, Fields: githubCommitFields()},
		{Name: "tags", Description: "Repository tags.", PrimaryKey: []string{"name"}, Fields: githubTagFields()},
		{Name: "releases", Description: "Repository releases.", PrimaryKey: []string{"id"}, CursorFields: []string{"published_at"}, Fields: githubReleaseFields()},
		{Name: "labels", Description: "Repository labels.", PrimaryKey: []string{"name"}, Fields: githubLabelFields()},
		{Name: "milestones", Description: "Repository milestones.", PrimaryKey: []string{"number"}, CursorFields: []string{"updated_at"}, Fields: githubMilestoneFields()},
		{Name: "issue_comments", Description: "Issue and pull request timeline comments at repository scope.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: githubIssueCommentFields()},
		{Name: "pull_request_review_comments", Description: "Pull request diff review comments.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: githubPullRequestReviewCommentFields()},
		{Name: "collaborators", Description: "Repository collaborators visible to the token.", PrimaryKey: []string{"id"}, Fields: githubUserFields()},
		{Name: "contributors", Description: "Repository contributors.", PrimaryKey: []string{"id"}, Fields: githubUserFields()},
		{Name: "stargazers", Description: "Users who starred the repository.", PrimaryKey: []string{"id"}, Fields: githubUserFields()},
		{Name: "subscribers", Description: "Users watching repository notifications.", PrimaryKey: []string{"id"}, Fields: githubUserFields()},
		{Name: "workflows", Description: "GitHub Actions workflows.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: githubWorkflowFields()},
		{Name: "workflow_runs", Description: "GitHub Actions workflow runs.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: githubWorkflowRunFields()},
		{Name: "workflow_artifacts", Description: "GitHub Actions workflow artifacts.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: githubWorkflowArtifactFields()},
		{Name: "deployments", Description: "Repository deployments.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: githubDeploymentFields()},
	}
}

func githubCommitQuery(cfg connectors.RuntimeConfig) url.Values {
	values := url.Values{}
	for _, key := range []string{"sha", "path", "author", "committer", "since", "until"} {
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			values.Set(key, value)
		}
	}
	return values
}

func githubWorkflowRunQuery(cfg connectors.RuntimeConfig) url.Values {
	values := url.Values{}
	for _, key := range []string{"actor", "branch", "event", "status", "created", "head_sha", "check_suite_id"} {
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			values.Set(key, value)
		}
	}
	return values
}

func githubRepositoryRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":        repository,
		"id":                item["id"],
		"node_id":           item["node_id"],
		"name":              item["name"],
		"full_name":         item["full_name"],
		"private":           item["private"],
		"description":       item["description"],
		"html_url":          item["html_url"],
		"default_branch":    item["default_branch"],
		"language":          item["language"],
		"stargazers_count":  item["stargazers_count"],
		"watchers_count":    item["watchers_count"],
		"forks_count":       item["forks_count"],
		"open_issues_count": item["open_issues_count"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"pushed_at":         item["pushed_at"],
	}
}

func githubBranchRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository": repository,
		"name":       item["name"],
		"protected":  item["protected"],
		"commit_sha": nestedString(item, "commit", "sha"),
		"commit_url": nestedString(item, "commit", "url"),
	}
}

func githubCommitRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":             repository,
		"sha":                    item["sha"],
		"node_id":                item["node_id"],
		"html_url":               item["html_url"],
		"url":                    item["url"],
		"author_login":           nestedString(item, "author", "login"),
		"author_id":              nestedValue(item, "author", "id"),
		"committer_login":        nestedString(item, "committer", "login"),
		"committer_id":           nestedValue(item, "committer", "id"),
		"commit_message":         nestedString(item, "commit", "message"),
		"commit_author_name":     nestedString(nestedMap(item, "commit"), "author", "name"),
		"commit_author_email":    nestedString(nestedMap(item, "commit"), "author", "email"),
		"commit_author_date":     nestedString(nestedMap(item, "commit"), "author", "date"),
		"commit_committer_name":  nestedString(nestedMap(item, "commit"), "committer", "name"),
		"commit_committer_email": nestedString(nestedMap(item, "commit"), "committer", "email"),
		"commit_committer_date":  nestedString(nestedMap(item, "commit"), "committer", "date"),
	}
}

func githubTagRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":  repository,
		"name":        item["name"],
		"zipball_url": item["zipball_url"],
		"tarball_url": item["tarball_url"],
		"commit_sha":  nestedString(item, "commit", "sha"),
		"commit_url":  nestedString(item, "commit", "url"),
		"node_id":     item["node_id"],
	}
}

func githubReleaseRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":       repository,
		"id":               item["id"],
		"node_id":          item["node_id"],
		"tag_name":         item["tag_name"],
		"target_commitish": item["target_commitish"],
		"name":             item["name"],
		"body":             item["body"],
		"draft":            item["draft"],
		"prerelease":       item["prerelease"],
		"html_url":         item["html_url"],
		"author_login":     nestedString(item, "author", "login"),
		"assets_count":     lenAnySlice(item["assets"]),
		"created_at":       item["created_at"],
		"published_at":     item["published_at"],
	}
}

func githubLabelRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":  repository,
		"id":          item["id"],
		"node_id":     item["node_id"],
		"name":        item["name"],
		"color":       item["color"],
		"description": item["description"],
		"default":     item["default"],
		"url":         item["url"],
	}
}

func githubMilestoneRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":    repository,
		"id":            item["id"],
		"node_id":       item["node_id"],
		"number":        item["number"],
		"state":         item["state"],
		"title":         item["title"],
		"description":   item["description"],
		"open_issues":   item["open_issues"],
		"closed_issues": item["closed_issues"],
		"creator_login": nestedString(item, "creator", "login"),
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"due_on":        item["due_on"],
		"closed_at":     item["closed_at"],
	}
}

func githubIssueCommentRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":         repository,
		"id":                 item["id"],
		"node_id":            item["node_id"],
		"issue_url":          item["issue_url"],
		"html_url":           item["html_url"],
		"body":               item["body"],
		"user_login":         nestedString(item, "user", "login"),
		"user_id":            nestedValue(item, "user", "id"),
		"author_association": item["author_association"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
	}
}

func githubPullRequestReviewCommentRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":             repository,
		"id":                     item["id"],
		"node_id":                item["node_id"],
		"pull_request_review_id": item["pull_request_review_id"],
		"diff_hunk":              item["diff_hunk"],
		"path":                   item["path"],
		"position":               item["position"],
		"original_position":      item["original_position"],
		"commit_id":              item["commit_id"],
		"original_commit_id":     item["original_commit_id"],
		"pull_request_url":       item["pull_request_url"],
		"html_url":               item["html_url"],
		"body":                   item["body"],
		"user_login":             nestedString(item, "user", "login"),
		"created_at":             item["created_at"],
		"updated_at":             item["updated_at"],
	}
}

func githubUserRecord(repository, relation string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":    repository,
		"relation":      relation,
		"login":         item["login"],
		"id":            item["id"],
		"node_id":       item["node_id"],
		"type":          item["type"],
		"site_admin":    item["site_admin"],
		"html_url":      item["html_url"],
		"contributions": item["contributions"],
		"role_name":     item["role_name"],
	}
}

func githubWorkflowRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository": repository,
		"id":         item["id"],
		"node_id":    item["node_id"],
		"name":       item["name"],
		"path":       item["path"],
		"state":      item["state"],
		"badge_url":  item["badge_url"],
		"html_url":   item["html_url"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func githubWorkflowRunRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":  repository,
		"id":          item["id"],
		"node_id":     item["node_id"],
		"name":        item["name"],
		"head_branch": item["head_branch"],
		"head_sha":    item["head_sha"],
		"status":      item["status"],
		"conclusion":  item["conclusion"],
		"event":       item["event"],
		"workflow_id": item["workflow_id"],
		"run_number":  item["run_number"],
		"run_attempt": item["run_attempt"],
		"html_url":    item["html_url"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func githubWorkflowArtifactRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":           repository,
		"id":                   item["id"],
		"node_id":              item["node_id"],
		"name":                 item["name"],
		"size_in_bytes":        item["size_in_bytes"],
		"url":                  item["url"],
		"archive_download_url": item["archive_download_url"],
		"expired":              item["expired"],
		"created_at":           item["created_at"],
		"updated_at":           item["updated_at"],
		"expires_at":           item["expires_at"],
		"workflow_run_id":      nestedValue(item, "workflow_run", "id"),
	}
}

func githubDeploymentRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":    repository,
		"id":            item["id"],
		"node_id":       item["node_id"],
		"sha":           item["sha"],
		"ref":           item["ref"],
		"task":          item["task"],
		"environment":   item["environment"],
		"description":   item["description"],
		"creator_login": nestedString(item, "creator", "login"),
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func nestedMap(item map[string]any, key string) map[string]any {
	nested, _ := item[key].(map[string]any)
	if nested == nil {
		return map[string]any{}
	}
	return nested
}

func githubRepositoryFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "full_name", Type: "string"}, {Name: "private", Type: "boolean"}, {Name: "description", Type: "string"}, {Name: "html_url", Type: "string"}, {Name: "default_branch", Type: "string"}, {Name: "language", Type: "string"}, {Name: "stargazers_count", Type: "integer"}, {Name: "watchers_count", Type: "integer"}, {Name: "forks_count", Type: "integer"}, {Name: "open_issues_count", Type: "integer"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}, {Name: "pushed_at", Type: "timestamp"}}
}

func githubBranchFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "name", Type: "string"}, {Name: "protected", Type: "boolean"}, {Name: "commit_sha", Type: "string"}, {Name: "commit_url", Type: "string"}}
}

func githubCommitFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "sha", Type: "string"}, {Name: "node_id", Type: "string"}, {Name: "html_url", Type: "string"}, {Name: "url", Type: "string"}, {Name: "author_login", Type: "string"}, {Name: "author_id", Type: "integer"}, {Name: "committer_login", Type: "string"}, {Name: "committer_id", Type: "integer"}, {Name: "commit_message", Type: "string"}, {Name: "commit_author_name", Type: "string"}, {Name: "commit_author_email", Type: "string"}, {Name: "commit_author_date", Type: "timestamp"}, {Name: "commit_committer_name", Type: "string"}, {Name: "commit_committer_email", Type: "string"}, {Name: "commit_committer_date", Type: "timestamp"}}
}

func githubTagFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "name", Type: "string"}, {Name: "zipball_url", Type: "string"}, {Name: "tarball_url", Type: "string"}, {Name: "commit_sha", Type: "string"}, {Name: "commit_url", Type: "string"}, {Name: "node_id", Type: "string"}}
}

func githubReleaseFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "tag_name", Type: "string"}, {Name: "target_commitish", Type: "string"}, {Name: "name", Type: "string"}, {Name: "body", Type: "string"}, {Name: "draft", Type: "boolean"}, {Name: "prerelease", Type: "boolean"}, {Name: "html_url", Type: "string"}, {Name: "author_login", Type: "string"}, {Name: "assets_count", Type: "integer"}, {Name: "created_at", Type: "timestamp"}, {Name: "published_at", Type: "timestamp"}}
}

func githubLabelFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "color", Type: "string"}, {Name: "description", Type: "string"}, {Name: "default", Type: "boolean"}, {Name: "url", Type: "string"}}
}

func githubMilestoneFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "number", Type: "integer"}, {Name: "state", Type: "string"}, {Name: "title", Type: "string"}, {Name: "description", Type: "string"}, {Name: "open_issues", Type: "integer"}, {Name: "closed_issues", Type: "integer"}, {Name: "creator_login", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}, {Name: "due_on", Type: "timestamp"}, {Name: "closed_at", Type: "timestamp"}}
}

func githubIssueCommentFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "issue_url", Type: "string"}, {Name: "html_url", Type: "string"}, {Name: "body", Type: "string"}, {Name: "user_login", Type: "string"}, {Name: "user_id", Type: "integer"}, {Name: "author_association", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}

func githubPullRequestReviewCommentFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "pull_request_review_id", Type: "integer"}, {Name: "diff_hunk", Type: "string"}, {Name: "path", Type: "string"}, {Name: "position", Type: "integer"}, {Name: "original_position", Type: "integer"}, {Name: "commit_id", Type: "string"}, {Name: "original_commit_id", Type: "string"}, {Name: "pull_request_url", Type: "string"}, {Name: "html_url", Type: "string"}, {Name: "body", Type: "string"}, {Name: "user_login", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}

func githubUserFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "relation", Type: "string"}, {Name: "login", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "site_admin", Type: "boolean"}, {Name: "html_url", Type: "string"}, {Name: "contributions", Type: "integer"}, {Name: "role_name", Type: "string"}}
}

func githubWorkflowFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "path", Type: "string"}, {Name: "state", Type: "string"}, {Name: "badge_url", Type: "string"}, {Name: "html_url", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}

func githubWorkflowRunFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "head_branch", Type: "string"}, {Name: "head_sha", Type: "string"}, {Name: "status", Type: "string"}, {Name: "conclusion", Type: "string"}, {Name: "event", Type: "string"}, {Name: "workflow_id", Type: "integer"}, {Name: "run_number", Type: "integer"}, {Name: "run_attempt", Type: "integer"}, {Name: "html_url", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}

func githubWorkflowArtifactFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "size_in_bytes", Type: "integer"}, {Name: "url", Type: "string"}, {Name: "archive_download_url", Type: "string"}, {Name: "expired", Type: "boolean"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}, {Name: "expires_at", Type: "timestamp"}, {Name: "workflow_run_id", Type: "integer"}}
}

func githubDeploymentFields() []connectors.Field {
	return []connectors.Field{{Name: "repository", Type: "string"}, {Name: "id", Type: "integer"}, {Name: "node_id", Type: "string"}, {Name: "sha", Type: "string"}, {Name: "ref", Type: "string"}, {Name: "task", Type: "string"}, {Name: "environment", Type: "string"}, {Name: "description", Type: "string"}, {Name: "creator_login", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}
