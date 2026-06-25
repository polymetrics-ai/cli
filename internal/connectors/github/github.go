// Package github implements the native pm GitHub connector as a per-system
// connector package. It self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics/internal/connectors"
)

const (
	githubDefaultBaseURL  = "https://api.github.com"
	githubDefaultPerPage  = 100
	githubDefaultMaxPages = 1
	githubAPIVersion      = "2026-03-10"
	githubUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("github", New)
}

// New returns the GitHub connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GitHub connector.
type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "github" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "github",
		DisplayName:     "GitHub",
		IntegrationType: "api",
		Description:     "Reads GitHub repository, issue, pull request, code, release, collaboration, and Actions data, and writes approved reverse ETL actions through the GitHub REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: true},
	}
}

func (g Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	owner, repo, err := githubRepository(cfg)
	if err != nil {
		return err
	}
	endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo)), nil)
	if err != nil {
		return err
	}
	var repository map[string]any
	if err := g.getJSON(ctx, cfg, endpoint, &repository); err != nil {
		return fmt.Errorf("check GitHub repository %s/%s: %w", owner, repo, err)
	}
	return nil
}

func (g Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	if _, _, err := githubRepository(cfg); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: g.Name(), Streams: githubStreams()}, nil
}

func (g Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	owner, repo, err := githubRepository(req.Config)
	if err != nil {
		return err
	}

	switch req.Stream {
	case "", "issues":
		values := githubBaseQuery(req.Config)
		if since := strings.TrimSpace(req.Config.Config["since"]); since != "" {
			values.Set("since", since)
		}
		path := fmt.Sprintf("/repos/%s/%s/issues", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, values, func(item map[string]any) (connectors.Record, bool) {
			if _, ok := item["pull_request"]; ok {
				return nil, false
			}
			return githubIssueRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "pull_requests":
		values := githubBaseQuery(req.Config)
		path := fmt.Sprintf("/repos/%s/%s/pulls", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, values, func(item map[string]any) (connectors.Record, bool) {
			return githubPullRequestRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "repository":
		path := fmt.Sprintf("/repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
		endpoint, err := githubEndpoint(req.Config, path, nil)
		if err != nil {
			return err
		}
		var item map[string]any
		if err := g.getJSON(ctx, req.Config, endpoint, &item); err != nil {
			return err
		}
		return emit(githubRepositoryRecord(req.Config.Config["repository"], item))
	case "branches":
		path := fmt.Sprintf("/repos/%s/%s/branches", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubBranchRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "commits":
		values := githubCommitQuery(req.Config)
		path := fmt.Sprintf("/repos/%s/%s/commits", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, values, func(item map[string]any) (connectors.Record, bool) {
			return githubCommitRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "tags":
		path := fmt.Sprintf("/repos/%s/%s/tags", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubTagRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "releases":
		path := fmt.Sprintf("/repos/%s/%s/releases", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubReleaseRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "labels":
		path := fmt.Sprintf("/repos/%s/%s/labels", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubLabelRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "milestones":
		values := githubBaseQuery(req.Config)
		path := fmt.Sprintf("/repos/%s/%s/milestones", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, values, func(item map[string]any) (connectors.Record, bool) {
			return githubMilestoneRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "issue_comments":
		values := url.Values{}
		if since := strings.TrimSpace(req.Config.Config["since"]); since != "" {
			values.Set("since", since)
		}
		path := fmt.Sprintf("/repos/%s/%s/issues/comments", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, values, func(item map[string]any) (connectors.Record, bool) {
			return githubIssueCommentRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "pull_request_review_comments":
		values := url.Values{}
		if since := strings.TrimSpace(req.Config.Config["since"]); since != "" {
			values.Set("since", since)
		}
		path := fmt.Sprintf("/repos/%s/%s/pulls/comments", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, values, func(item map[string]any) (connectors.Record, bool) {
			return githubPullRequestReviewCommentRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "collaborators":
		path := fmt.Sprintf("/repos/%s/%s/collaborators", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubUserRecord(req.Config.Config["repository"], "collaborator", item), true
		}, emit)
	case "contributors":
		path := fmt.Sprintf("/repos/%s/%s/contributors", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubUserRecord(req.Config.Config["repository"], "contributor", item), true
		}, emit)
	case "stargazers":
		path := fmt.Sprintf("/repos/%s/%s/stargazers", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubUserRecord(req.Config.Config["repository"], "stargazer", item), true
		}, emit)
	case "subscribers":
		path := fmt.Sprintf("/repos/%s/%s/subscribers", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubUserRecord(req.Config.Config["repository"], "subscriber", item), true
		}, emit)
	case "workflows":
		path := fmt.Sprintf("/repos/%s/%s/actions/workflows", url.PathEscape(owner), url.PathEscape(repo))
		return g.readEnvelopePages(ctx, req.Config, path, url.Values{}, "workflows", func(item map[string]any) (connectors.Record, bool) {
			return githubWorkflowRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "workflow_runs":
		values := githubWorkflowRunQuery(req.Config)
		path := fmt.Sprintf("/repos/%s/%s/actions/runs", url.PathEscape(owner), url.PathEscape(repo))
		return g.readEnvelopePages(ctx, req.Config, path, values, "workflow_runs", func(item map[string]any) (connectors.Record, bool) {
			return githubWorkflowRunRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "workflow_artifacts":
		path := fmt.Sprintf("/repos/%s/%s/actions/artifacts", url.PathEscape(owner), url.PathEscape(repo))
		return g.readEnvelopePages(ctx, req.Config, path, url.Values{}, "artifacts", func(item map[string]any) (connectors.Record, bool) {
			return githubWorkflowArtifactRecord(req.Config.Config["repository"], item), true
		}, emit)
	case "deployments":
		path := fmt.Sprintf("/repos/%s/%s/deployments", url.PathEscape(owner), url.PathEscape(repo))
		return g.readPages(ctx, req.Config, path, url.Values{}, func(item map[string]any) (connectors.Record, bool) {
			return githubDeploymentRecord(req.Config.Config["repository"], item), true
		}, emit)
	default:
		return fmt.Errorf("github stream %q not found", req.Stream)
	}
}

func (g Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, _, err := githubRepository(req.Config); err != nil {
		return err
	}
	action, err := githubNormalizeWriteAction(req.Action)
	if err != nil {
		return err
	}
	if !githubHasWriteAuth(req.Config) {
		return fmt.Errorf("github write action %q requires token auth or github_app auth", action)
	}
	for i, record := range records {
		if err := githubValidateWriteRecord(action, record); err != nil {
			return fmt.Errorf("github %s record %d: %w", action, i+1, err)
		}
	}
	return nil
}

func (g Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := g.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	owner, repo, _ := githubRepository(req.Config)
	action, _ := githubNormalizeWriteAction(req.Action)
	result := connectors.WriteResult{}
	for i, record := range records {
		if err := g.executeWriteAction(ctx, req.Config, owner, repo, action, record); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("github %s record %d: %w", action, i+1, err)
		}
		result.RecordsWritten++
	}
	return result, nil
}

func (g Connector) readPages(ctx context.Context, cfg connectors.RuntimeConfig, path string, values url.Values, normalize func(map[string]any) (connectors.Record, bool), emit func(connectors.Record) error) error {
	perPage, err := githubPositiveInt(cfg.Config["per_page"], githubDefaultPerPage, 1, 100, "per_page")
	if err != nil {
		return err
	}
	maxPages, err := githubMaxPages(cfg.Config["max_pages"])
	if err != nil {
		return err
	}
	values.Set("per_page", strconv.Itoa(perPage))

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		values.Set("page", strconv.Itoa(page))
		endpoint, err := githubEndpoint(cfg, path, values)
		if err != nil {
			return err
		}
		var pageRecords []map[string]any
		if err := g.getJSON(ctx, cfg, endpoint, &pageRecords); err != nil {
			return fmt.Errorf("read GitHub page %d: %w", page, err)
		}
		if len(pageRecords) == 0 {
			return nil
		}
		for _, item := range pageRecords {
			record, ok := normalize(item)
			if !ok {
				continue
			}
			if err := emit(record); err != nil {
				return err
			}
		}
		if len(pageRecords) < perPage {
			return nil
		}
	}
	return nil
}

func (g Connector) readEnvelopePages(ctx context.Context, cfg connectors.RuntimeConfig, path string, values url.Values, key string, normalize func(map[string]any) (connectors.Record, bool), emit func(connectors.Record) error) error {
	perPage, err := githubPositiveInt(cfg.Config["per_page"], githubDefaultPerPage, 1, 100, "per_page")
	if err != nil {
		return err
	}
	maxPages, err := githubMaxPages(cfg.Config["max_pages"])
	if err != nil {
		return err
	}
	values.Set("per_page", strconv.Itoa(perPage))

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		values.Set("page", strconv.Itoa(page))
		endpoint, err := githubEndpoint(cfg, path, values)
		if err != nil {
			return err
		}
		var envelope map[string]any
		if err := g.getJSON(ctx, cfg, endpoint, &envelope); err != nil {
			return fmt.Errorf("read GitHub page %d: %w", page, err)
		}
		items, _ := envelope[key].([]any)
		if len(items) == 0 {
			return nil
		}
		for _, raw := range items {
			item, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			record, ok := normalize(item)
			if !ok {
				continue
			}
			if err := emit(record); err != nil {
				return err
			}
		}
		if len(items) < perPage {
			return nil
		}
	}
	return nil
}

func (g Connector) getJSON(ctx context.Context, cfg connectors.RuntimeConfig, endpoint string, out any) error {
	return g.doJSON(ctx, cfg, http.MethodGet, endpoint, nil, out)
}

func (g Connector) doJSON(ctx context.Context, cfg connectors.RuntimeConfig, method, endpoint string, payload any, out any) error {
	authHeader, err := g.authorizationHeader(ctx, cfg)
	if err != nil {
		return err
	}
	return g.doJSONWithAuth(ctx, cfg, method, endpoint, payload, out, authHeader)
}

func (g Connector) doJSONWithAuth(ctx context.Context, cfg connectors.RuntimeConfig, method, endpoint string, payload any, out any, authHeader string) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode GitHub request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("build GitHub request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)
	req.Header.Set("User-Agent", githubUserAgent)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := g.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("send GitHub request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("GitHub API returned %s: %s", resp.Status, msg)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}

func (g Connector) httpClient() *http.Client {
	if g.Client != nil {
		return g.Client
	}
	return &http.Client{Timeout: 20 * time.Second}
}

func (g Connector) executeWriteAction(ctx context.Context, cfg connectors.RuntimeConfig, owner, repo, action string, record connectors.Record) error {
	switch action {
	case "create_issue":
		payload, err := githubCreateIssuePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/issues", url.PathEscape(owner), url.PathEscape(repo)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "update_issue":
		number, err := githubRequiredNumber(record, "issue_number", "number")
		if err != nil {
			return err
		}
		payload, err := githubUpdateIssuePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil)
	case "comment_issue":
		number, err := githubRequiredNumber(record, "issue_number", "pull_number", "number")
		if err != nil {
			return err
		}
		payload, err := githubCommentPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/issues/%d/comments", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "close_issue":
		number, err := githubRequiredNumber(record, "issue_number", "number")
		if err != nil {
			return err
		}
		if comment := githubOptionalString(record, "comment"); comment != "" {
			if err := g.writeIssueComment(ctx, cfg, owner, repo, number, comment); err != nil {
				return err
			}
		}
		payload := map[string]any{"state": "closed"}
		if reason := githubOptionalString(record, "state_reason"); reason != "" {
			payload["state_reason"] = reason
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil)
	case "create_pull_request":
		payload, err := githubCreatePullRequestPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/pulls", url.PathEscape(owner), url.PathEscape(repo)), nil)
		if err != nil {
			return err
		}
		var response map[string]any
		if err := g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, &response); err != nil {
			return err
		}
		number, err := githubResponseNumber(response)
		if err != nil {
			return err
		}
		return g.writePullRequestFollowups(ctx, cfg, owner, repo, number, record)
	case "update_pull_request":
		number, err := githubRequiredNumber(record, "pull_number", "number")
		if err != nil {
			return err
		}
		payload, err := githubUpdatePullRequestPayload(record)
		if err != nil {
			return err
		}
		if len(payload) > 0 {
			endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/pulls/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
			if err != nil {
				return err
			}
			if err := g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil); err != nil {
				return err
			}
		}
		return g.writePullRequestFollowups(ctx, cfg, owner, repo, number, record)
	case "close_pull_request":
		number, err := githubRequiredNumber(record, "pull_number", "number")
		if err != nil {
			return err
		}
		if comment := githubOptionalString(record, "comment"); comment != "" {
			if err := g.writeIssueComment(ctx, cfg, owner, repo, number, comment); err != nil {
				return err
			}
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/pulls/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPatch, endpoint, map[string]any{"state": "closed"}, nil)
	case "request_reviewers":
		number, err := githubRequiredNumber(record, "pull_number", "number")
		if err != nil {
			return err
		}
		return g.writeReviewers(ctx, cfg, owner, repo, number, record)
	case "merge_pull_request":
		number, err := githubRequiredNumber(record, "pull_number", "number")
		if err != nil {
			return err
		}
		payload, err := githubMergePullRequestPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPut, endpoint, payload, nil)
	case "create_label":
		payload, err := githubCreateLabelPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/labels", url.PathEscape(owner), url.PathEscape(repo)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "update_label":
		name, err := githubRequiredString(record, "name")
		if err != nil {
			return err
		}
		payload, err := githubUpdateLabelPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/labels/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(name)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil)
	case "delete_label":
		name, err := githubRequiredString(record, "name")
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/labels/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(name)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodDelete, endpoint, nil, nil)
	case "create_milestone":
		payload, err := githubCreateMilestonePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/milestones", url.PathEscape(owner), url.PathEscape(repo)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "update_milestone":
		number, err := githubRequiredNumber(record, "milestone_number", "number")
		if err != nil {
			return err
		}
		payload, err := githubUpdateMilestonePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/milestones/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil)
	case "delete_milestone":
		number, err := githubRequiredNumber(record, "milestone_number", "number")
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/milestones/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodDelete, endpoint, nil, nil)
	case "create_release":
		payload, err := githubCreateReleasePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/releases", url.PathEscape(owner), url.PathEscape(repo)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "update_release":
		id, err := githubRequiredNumber(record, "release_id", "id")
		if err != nil {
			return err
		}
		payload, err := githubUpdateReleasePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/releases/%d", url.PathEscape(owner), url.PathEscape(repo), id), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil)
	case "delete_release":
		id, err := githubRequiredNumber(record, "release_id", "id")
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/releases/%d", url.PathEscape(owner), url.PathEscape(repo), id), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodDelete, endpoint, nil, nil)
	case "dispatch_workflow":
		workflowID, err := githubRequiredString(record, "workflow_id")
		if err != nil {
			return err
		}
		payload, err := githubDispatchWorkflowPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/actions/workflows/%s/dispatches", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(workflowID)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "rerun_workflow_run":
		id, err := githubRequiredNumber(record, "run_id", "workflow_run_id", "id")
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/actions/runs/%d/rerun", url.PathEscape(owner), url.PathEscape(repo), id), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, nil, nil)
	case "cancel_workflow_run":
		id, err := githubRequiredNumber(record, "run_id", "workflow_run_id", "id")
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/actions/runs/%d/cancel", url.PathEscape(owner), url.PathEscape(repo), id), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, nil, nil)
	case "delete_workflow_run":
		id, err := githubRequiredNumber(record, "run_id", "workflow_run_id", "id")
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/actions/runs/%d", url.PathEscape(owner), url.PathEscape(repo), id), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodDelete, endpoint, nil, nil)
	case "create_pull_request_review":
		number, err := githubRequiredNumber(record, "pull_number", "number")
		if err != nil {
			return err
		}
		payload, err := githubCreatePullRequestReviewPayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
	case "create_or_update_file":
		path, err := githubRequiredString(record, "path")
		if err != nil {
			return err
		}
		payload, err := githubCreateOrUpdateFilePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/contents/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(path)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodPut, endpoint, payload, nil)
	case "delete_file":
		path, err := githubRequiredString(record, "path")
		if err != nil {
			return err
		}
		payload, err := githubDeleteFilePayload(record)
		if err != nil {
			return err
		}
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/contents/%s", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(path)), nil)
		if err != nil {
			return err
		}
		return g.doJSON(ctx, cfg, http.MethodDelete, endpoint, payload, nil)
	default:
		return fmt.Errorf("unsupported github write action %q", action)
	}
}

func (g Connector) writePullRequestFollowups(ctx context.Context, cfg connectors.RuntimeConfig, owner, repo string, number int, record connectors.Record) error {
	if payload := githubIssueMetadataPayload(record); len(payload) > 0 {
		endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(repo), number), nil)
		if err != nil {
			return err
		}
		if err := g.doJSON(ctx, cfg, http.MethodPatch, endpoint, payload, nil); err != nil {
			return err
		}
	}
	return g.writeReviewers(ctx, cfg, owner, repo, number, record)
}

func (g Connector) writeReviewers(ctx context.Context, cfg connectors.RuntimeConfig, owner, repo string, number int, record connectors.Record) error {
	payload := githubReviewersPayload(record)
	if len(payload) == 0 {
		return nil
	}
	endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/pulls/%d/requested_reviewers", url.PathEscape(owner), url.PathEscape(repo), number), nil)
	if err != nil {
		return err
	}
	return g.doJSON(ctx, cfg, http.MethodPost, endpoint, payload, nil)
}

func (g Connector) writeIssueComment(ctx context.Context, cfg connectors.RuntimeConfig, owner, repo string, number int, body string) error {
	endpoint, err := githubEndpoint(cfg, fmt.Sprintf("/repos/%s/%s/issues/%d/comments", url.PathEscape(owner), url.PathEscape(repo), number), nil)
	if err != nil {
		return err
	}
	return g.doJSON(ctx, cfg, http.MethodPost, endpoint, map[string]any{"body": body}, nil)
}

func githubNormalizeWriteAction(raw string) (string, error) {
	action := strings.ToLower(strings.TrimSpace(raw))
	action = strings.ReplaceAll(action, "-", "_")
	action = strings.ReplaceAll(action, " ", "_")
	switch action {
	case "create_issue", "issue_create", "new_issue":
		return "create_issue", nil
	case "update_issue", "edit_issue", "issue_update", "issue_edit":
		return "update_issue", nil
	case "comment_issue", "create_issue_comment", "issue_comment", "comment_pr", "pr_comment", "create_pr_comment":
		return "comment_issue", nil
	case "close_issue":
		return "close_issue", nil
	case "create_pull_request", "create_pr", "pull_request_create", "pr_create", "new_pr":
		return "create_pull_request", nil
	case "update_pull_request", "update_pr", "edit_pull_request", "edit_pr", "pull_request_update", "pr_update":
		return "update_pull_request", nil
	case "close_pull_request", "close_pr":
		return "close_pull_request", nil
	case "request_reviewers", "add_reviewers", "request_pr_reviewers":
		return "request_reviewers", nil
	case "merge_pull_request", "merge_pr", "pr_merge":
		return "merge_pull_request", nil
	case "create_label", "label_create", "new_label":
		return "create_label", nil
	case "update_label", "edit_label", "label_update", "label_edit":
		return "update_label", nil
	case "delete_label", "remove_label", "label_delete":
		return "delete_label", nil
	case "create_milestone", "milestone_create", "new_milestone":
		return "create_milestone", nil
	case "update_milestone", "edit_milestone", "milestone_update", "milestone_edit":
		return "update_milestone", nil
	case "delete_milestone", "remove_milestone", "milestone_delete":
		return "delete_milestone", nil
	case "create_release", "release_create", "new_release":
		return "create_release", nil
	case "update_release", "edit_release", "release_update", "release_edit":
		return "update_release", nil
	case "delete_release", "remove_release", "release_delete":
		return "delete_release", nil
	case "dispatch_workflow", "trigger_workflow", "workflow_dispatch", "run_workflow":
		return "dispatch_workflow", nil
	case "rerun_workflow_run", "rerun_run", "workflow_run_rerun":
		return "rerun_workflow_run", nil
	case "cancel_workflow_run", "cancel_run", "workflow_run_cancel":
		return "cancel_workflow_run", nil
	case "delete_workflow_run", "remove_workflow_run", "workflow_run_delete":
		return "delete_workflow_run", nil
	case "create_pull_request_review", "create_pr_review", "pull_request_review", "review_pull_request", "review_pr":
		return "create_pull_request_review", nil
	case "create_or_update_file", "upsert_file", "put_file", "create_file", "update_file":
		return "create_or_update_file", nil
	case "delete_file", "remove_file":
		return "delete_file", nil
	default:
		return "", fmt.Errorf("unsupported github write action %q", raw)
	}
}

func githubValidateWriteRecord(action string, record connectors.Record) error {
	switch action {
	case "create_issue":
		_, err := githubCreateIssuePayload(record)
		return err
	case "update_issue":
		if _, err := githubRequiredNumber(record, "issue_number", "number"); err != nil {
			return err
		}
		_, err := githubUpdateIssuePayload(record)
		return err
	case "comment_issue":
		if _, err := githubRequiredNumber(record, "issue_number", "pull_number", "number"); err != nil {
			return err
		}
		_, err := githubCommentPayload(record)
		return err
	case "close_issue":
		_, err := githubRequiredNumber(record, "issue_number", "number")
		return err
	case "create_pull_request":
		_, err := githubCreatePullRequestPayload(record)
		return err
	case "update_pull_request":
		if _, err := githubRequiredNumber(record, "pull_number", "number"); err != nil {
			return err
		}
		core, err := githubUpdatePullRequestPayload(record)
		if err != nil {
			return err
		}
		if len(core) == 0 && len(githubIssueMetadataPayload(record)) == 0 && len(githubReviewersPayload(record)) == 0 {
			return errors.New("update_pull_request requires at least one mutable field")
		}
		return nil
	case "close_pull_request":
		_, err := githubRequiredNumber(record, "pull_number", "number")
		return err
	case "request_reviewers":
		if _, err := githubRequiredNumber(record, "pull_number", "number"); err != nil {
			return err
		}
		if len(githubReviewersPayload(record)) == 0 {
			return errors.New("request_reviewers requires reviewers or team_reviewers")
		}
		return nil
	case "merge_pull_request":
		if _, err := githubRequiredNumber(record, "pull_number", "number"); err != nil {
			return err
		}
		_, err := githubMergePullRequestPayload(record)
		return err
	case "create_label":
		_, err := githubCreateLabelPayload(record)
		return err
	case "update_label":
		if _, err := githubRequiredString(record, "name"); err != nil {
			return err
		}
		_, err := githubUpdateLabelPayload(record)
		return err
	case "delete_label":
		_, err := githubRequiredString(record, "name")
		return err
	case "create_milestone":
		_, err := githubCreateMilestonePayload(record)
		return err
	case "update_milestone":
		if _, err := githubRequiredNumber(record, "milestone_number", "number"); err != nil {
			return err
		}
		_, err := githubUpdateMilestonePayload(record)
		return err
	case "delete_milestone":
		_, err := githubRequiredNumber(record, "milestone_number", "number")
		return err
	case "create_release":
		_, err := githubCreateReleasePayload(record)
		return err
	case "update_release":
		if _, err := githubRequiredNumber(record, "release_id", "id"); err != nil {
			return err
		}
		_, err := githubUpdateReleasePayload(record)
		return err
	case "delete_release":
		_, err := githubRequiredNumber(record, "release_id", "id")
		return err
	case "dispatch_workflow":
		if _, err := githubRequiredString(record, "workflow_id"); err != nil {
			return err
		}
		_, err := githubDispatchWorkflowPayload(record)
		return err
	case "rerun_workflow_run", "cancel_workflow_run", "delete_workflow_run":
		_, err := githubRequiredNumber(record, "run_id", "workflow_run_id", "id")
		return err
	case "create_pull_request_review":
		if _, err := githubRequiredNumber(record, "pull_number", "number"); err != nil {
			return err
		}
		_, err := githubCreatePullRequestReviewPayload(record)
		return err
	case "create_or_update_file":
		if _, err := githubRequiredString(record, "path"); err != nil {
			return err
		}
		_, err := githubCreateOrUpdateFilePayload(record)
		return err
	case "delete_file":
		if _, err := githubRequiredString(record, "path"); err != nil {
			return err
		}
		_, err := githubDeleteFilePayload(record)
		return err
	default:
		return fmt.Errorf("unsupported github write action %q", action)
	}
}

func githubCreateIssuePayload(record connectors.Record) (map[string]any, error) {
	title, err := githubRequiredString(record, "title")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{"title": title}
	if body := githubOptionalString(record, "body"); body != "" {
		payload["body"] = body
	}
	for key, value := range githubIssueMetadataPayload(record) {
		payload[key] = value
	}
	if issueType := githubOptionalString(record, "type"); issueType != "" {
		payload["type"] = issueType
	}
	return payload, nil
}

func githubUpdateIssuePayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if title := githubOptionalString(record, "title"); title != "" {
		payload["title"] = title
	}
	if body := githubOptionalString(record, "body"); body != "" {
		payload["body"] = body
	}
	if state := githubOptionalString(record, "state"); state != "" {
		if state != "open" && state != "closed" {
			return nil, fmt.Errorf("state must be open or closed, got %q", state)
		}
		payload["state"] = state
	}
	if reason := githubOptionalString(record, "state_reason"); reason != "" {
		payload["state_reason"] = reason
	}
	if issueType := githubOptionalString(record, "type"); issueType != "" {
		payload["type"] = issueType
	}
	for key, value := range githubIssueMetadataPayload(record) {
		payload[key] = value
	}
	if len(payload) == 0 {
		return nil, errors.New("update_issue requires at least one mutable field")
	}
	return payload, nil
}

func githubIssueMetadataPayload(record connectors.Record) map[string]any {
	payload := map[string]any{}
	if labels := githubStringSlice(record, "labels"); len(labels) > 0 {
		payload["labels"] = labels
	}
	if assignees := githubStringSlice(record, "assignees"); len(assignees) > 0 {
		payload["assignees"] = assignees
	}
	if milestone, ok := githubOptionalNumber(record, "milestone"); ok {
		payload["milestone"] = milestone
	}
	return payload
}

func githubCommentPayload(record connectors.Record) (map[string]any, error) {
	body, err := githubRequiredString(record, "body")
	if err != nil {
		return nil, err
	}
	return map[string]any{"body": body}, nil
}

func githubCreatePullRequestPayload(record connectors.Record) (map[string]any, error) {
	head, headErr := githubRequiredString(record, "head")
	base, baseErr := githubRequiredString(record, "base")
	if headErr != nil || baseErr != nil {
		return nil, errors.Join(headErr, baseErr)
	}
	payload := map[string]any{"head": head, "base": base}
	if issue, ok := githubOptionalNumber(record, "issue"); ok {
		payload["issue"] = issue
	} else {
		title, err := githubRequiredString(record, "title")
		if err != nil {
			return nil, err
		}
		payload["title"] = title
		if body := githubOptionalString(record, "body"); body != "" {
			payload["body"] = body
		}
	}
	if draft, ok, err := githubOptionalBool(record, "draft"); err != nil {
		return nil, err
	} else if ok {
		payload["draft"] = draft
	}
	if maintainers, ok, err := githubOptionalBool(record, "maintainer_can_modify"); err != nil {
		return nil, err
	} else if ok {
		payload["maintainer_can_modify"] = maintainers
	}
	return payload, nil
}

func githubUpdatePullRequestPayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if title := githubOptionalString(record, "title"); title != "" {
		payload["title"] = title
	}
	if body := githubOptionalString(record, "body"); body != "" {
		payload["body"] = body
	}
	if state := githubOptionalString(record, "state"); state != "" {
		if state != "open" && state != "closed" {
			return nil, fmt.Errorf("state must be open or closed, got %q", state)
		}
		payload["state"] = state
	}
	if base := githubOptionalString(record, "base"); base != "" {
		payload["base"] = base
	}
	if maintainers, ok, err := githubOptionalBool(record, "maintainer_can_modify"); err != nil {
		return nil, err
	} else if ok {
		payload["maintainer_can_modify"] = maintainers
	}
	return payload, nil
}

func githubReviewersPayload(record connectors.Record) map[string]any {
	payload := map[string]any{}
	if reviewers := githubStringSlice(record, "reviewers"); len(reviewers) > 0 {
		payload["reviewers"] = reviewers
	}
	if teams := githubStringSlice(record, "team_reviewers"); len(teams) > 0 {
		payload["team_reviewers"] = teams
	}
	return payload
}

func githubMergePullRequestPayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if title := githubOptionalString(record, "commit_title"); title != "" {
		payload["commit_title"] = title
	}
	if message := githubOptionalString(record, "commit_message"); message != "" {
		payload["commit_message"] = message
	}
	if sha := githubOptionalString(record, "sha"); sha != "" {
		payload["sha"] = sha
	}
	if method := githubOptionalString(record, "merge_method"); method != "" {
		switch method {
		case "merge", "squash", "rebase":
			payload["merge_method"] = method
		default:
			return nil, fmt.Errorf("merge_method must be merge, squash, or rebase, got %q", method)
		}
	}
	return payload, nil
}

func githubCreateLabelPayload(record connectors.Record) (map[string]any, error) {
	name, err := githubRequiredString(record, "name")
	if err != nil {
		return nil, err
	}
	color, err := githubRequiredString(record, "color")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{"name": name, "color": strings.TrimPrefix(color, "#")}
	if description := githubOptionalString(record, "description"); description != "" {
		payload["description"] = description
	}
	return payload, nil
}

func githubUpdateLabelPayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if newName := githubOptionalString(record, "new_name"); newName != "" {
		payload["new_name"] = newName
	}
	if color := githubOptionalString(record, "color"); color != "" {
		payload["color"] = strings.TrimPrefix(color, "#")
	}
	if description := githubOptionalString(record, "description"); description != "" {
		payload["description"] = description
	}
	if len(payload) == 0 {
		return nil, errors.New("update_label requires new_name, color, or description")
	}
	return payload, nil
}

func githubCreateMilestonePayload(record connectors.Record) (map[string]any, error) {
	title, err := githubRequiredString(record, "title")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{"title": title}
	if err := githubApplyMilestoneFields(payload, record); err != nil {
		return nil, err
	}
	return payload, nil
}

func githubUpdateMilestonePayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if title := githubOptionalString(record, "title"); title != "" {
		payload["title"] = title
	}
	if err := githubApplyMilestoneFields(payload, record); err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, errors.New("update_milestone requires title, state, description, or due_on")
	}
	return payload, nil
}

func githubApplyMilestoneFields(payload map[string]any, record connectors.Record) error {
	if state := githubOptionalString(record, "state"); state != "" {
		if state != "open" && state != "closed" {
			return fmt.Errorf("state must be open or closed, got %q", state)
		}
		payload["state"] = state
	}
	if description := githubOptionalString(record, "description"); description != "" {
		payload["description"] = description
	}
	if dueOn := githubOptionalString(record, "due_on"); dueOn != "" {
		payload["due_on"] = dueOn
	}
	return nil
}

func githubCreateReleasePayload(record connectors.Record) (map[string]any, error) {
	tagName, err := githubRequiredString(record, "tag_name")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{"tag_name": tagName}
	if err := githubApplyReleaseFields(payload, record); err != nil {
		return nil, err
	}
	return payload, nil
}

func githubUpdateReleasePayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if tagName := githubOptionalString(record, "tag_name"); tagName != "" {
		payload["tag_name"] = tagName
	}
	if err := githubApplyReleaseFields(payload, record); err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, errors.New("update_release requires at least one mutable field")
	}
	return payload, nil
}

func githubApplyReleaseFields(payload map[string]any, record connectors.Record) error {
	for _, key := range []string{"target_commitish", "name", "body"} {
		if value := githubOptionalString(record, key); value != "" {
			payload[key] = value
		}
	}
	for _, key := range []string{"draft", "prerelease", "generate_release_notes", "make_latest"} {
		if value, ok, err := githubOptionalBool(record, key); err != nil {
			return err
		} else if ok {
			payload[key] = value
		}
	}
	return nil
}

func githubDispatchWorkflowPayload(record connectors.Record) (map[string]any, error) {
	ref, err := githubRequiredString(record, "ref")
	if err != nil {
		return nil, err
	}
	payload := map[string]any{"ref": ref}
	inputs, err := githubOptionalObject(record, "inputs")
	if err != nil {
		return nil, err
	}
	if len(inputs) > 0 {
		payload["inputs"] = inputs
	}
	return payload, nil
}

func githubCreatePullRequestReviewPayload(record connectors.Record) (map[string]any, error) {
	payload := map[string]any{}
	if body := githubOptionalString(record, "body"); body != "" {
		payload["body"] = body
	}
	if commitID := githubOptionalString(record, "commit_id"); commitID != "" {
		payload["commit_id"] = commitID
	}
	if event := githubOptionalString(record, "event"); event != "" {
		normalized := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(event, "-", "_"), " ", "_"))
		switch normalized {
		case "APPROVE", "REQUEST_CHANGES", "COMMENT":
			payload["event"] = normalized
		default:
			return nil, fmt.Errorf("event must be APPROVE, REQUEST_CHANGES, or COMMENT, got %q", event)
		}
	}
	comments, err := githubOptionalArray(record, "comments")
	if err != nil {
		return nil, err
	}
	if len(comments) > 0 {
		payload["comments"] = comments
	}
	return payload, nil
}

func githubCreateOrUpdateFilePayload(record connectors.Record) (map[string]any, error) {
	message, err := githubRequiredString(record, "message")
	if err != nil {
		return nil, err
	}
	content := githubOptionalString(record, "content_base64")
	if content == "" {
		raw, err := githubRequiredString(record, "content")
		if err != nil {
			return nil, errors.Join(err, errors.New("content_base64 can be supplied instead"))
		}
		content = base64.StdEncoding.EncodeToString([]byte(raw))
	}
	payload := map[string]any{"message": message, "content": content}
	if sha := githubOptionalString(record, "sha"); sha != "" {
		payload["sha"] = sha
	}
	if err := githubApplyContentCommitFields(payload, record); err != nil {
		return nil, err
	}
	return payload, nil
}

func githubDeleteFilePayload(record connectors.Record) (map[string]any, error) {
	message, msgErr := githubRequiredString(record, "message")
	sha, shaErr := githubRequiredString(record, "sha")
	if msgErr != nil || shaErr != nil {
		return nil, errors.Join(msgErr, shaErr)
	}
	payload := map[string]any{"message": message, "sha": sha}
	if err := githubApplyContentCommitFields(payload, record); err != nil {
		return nil, err
	}
	return payload, nil
}

func githubApplyContentCommitFields(payload map[string]any, record connectors.Record) error {
	if branch := githubOptionalString(record, "branch"); branch != "" {
		payload["branch"] = branch
	}
	for _, key := range []string{"committer", "author"} {
		object, err := githubOptionalObject(record, key)
		if err != nil {
			return err
		}
		if len(object) > 0 {
			payload[key] = object
		}
	}
	return nil
}

func githubRequiredString(record connectors.Record, key string) (string, error) {
	value := githubOptionalString(record, key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func githubOptionalString(record connectors.Record, key string) string {
	value, ok := record[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	case json.Number:
		return typed.String()
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func githubRequiredNumber(record connectors.Record, keys ...string) (int, error) {
	for _, key := range keys {
		if value, ok := githubOptionalNumber(record, key); ok {
			return value, nil
		}
	}
	return 0, fmt.Errorf("%s is required", strings.Join(keys, " or "))
}

func githubOptionalNumber(record connectors.Record, key string) (int, bool) {
	value, ok := record[key]
	if !ok || value == nil {
		return 0, false
	}
	number, ok := githubAnyInt(value)
	return number, ok
}

func githubAnyInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		if typed == float64(int(typed)) {
			return int(typed), true
		}
	case json.Number:
		if n, err := typed.Int64(); err == nil {
			return int(n), true
		}
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		if n, err := strconv.Atoi(trimmed); err == nil {
			return n, true
		}
	}
	return 0, false
}

func githubOptionalBool(record connectors.Record, key string) (bool, bool, error) {
	value, ok := record[key]
	if !ok || value == nil {
		return false, false, nil
	}
	switch typed := value.(type) {
	case bool:
		return typed, true, nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return false, false, nil
		}
		parsed, err := strconv.ParseBool(trimmed)
		if err != nil {
			return false, false, fmt.Errorf("%s must be boolean: %w", key, err)
		}
		return parsed, true, nil
	default:
		return false, false, fmt.Errorf("%s must be boolean", key)
	}
}

func githubStringSlice(record connectors.Record, key string) []string {
	value, ok := record[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return compactStrings(typed)
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			values = append(values, githubAnyString(item))
		}
		return compactStrings(values)
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil
		}
		if strings.HasPrefix(trimmed, "[") {
			var values []string
			if err := json.Unmarshal([]byte(trimmed), &values); err == nil {
				return compactStrings(values)
			}
		}
		return compactStrings(strings.Split(trimmed, ","))
	default:
		return compactStrings([]string{githubAnyString(typed)})
	}
}

func githubOptionalObject(record connectors.Record, key string) (map[string]any, error) {
	value, ok := record[key]
	if !ok || value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed, nil
	case connectors.Record:
		out := map[string]any{}
		for k, v := range typed {
			out[k] = v
		}
		return out, nil
	case map[string]string:
		out := map[string]any{}
		for k, v := range typed {
			out[k] = v
		}
		return out, nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil, nil
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
			return nil, fmt.Errorf("%s must be a JSON object: %w", key, err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s must be an object or JSON object string", key)
	}
}

func githubOptionalArray(record connectors.Record, key string) ([]any, error) {
	value, ok := record[key]
	if !ok || value == nil {
		return nil, nil
	}
	switch typed := value.(type) {
	case []any:
		return typed, nil
	case []map[string]any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out, nil
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out, nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return nil, nil
		}
		var out []any
		if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
			return nil, fmt.Errorf("%s must be a JSON array: %w", key, err)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s must be an array or JSON array string", key)
	}
}

func githubAnyString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func githubResponseNumber(response map[string]any) (int, error) {
	if number, ok := githubAnyInt(response["number"]); ok {
		return number, nil
	}
	return 0, errors.New("GitHub response missing number")
}

func githubRepository(cfg connectors.RuntimeConfig) (string, string, error) {
	repository := strings.TrimSpace(cfg.Config["repository"])
	if repository == "" {
		repository = strings.TrimSpace(cfg.Config["repo"])
	}
	owner, repo, ok := strings.Cut(repository, "/")
	if !ok || owner == "" || repo == "" || strings.Contains(repo, "/") {
		return "", "", errors.New(`github connector requires config repository in "owner/repo" format`)
	}
	return owner, repo, nil
}

func githubEndpoint(cfg connectors.RuntimeConfig, path string, values url.Values) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = githubDefaultBaseURL
	}
	base = strings.TrimRight(base, "/")
	parsed, err := url.Parse(base + path)
	if err != nil {
		return "", fmt.Errorf("build GitHub endpoint: %w", err)
	}
	if values != nil {
		parsed.RawQuery = values.Encode()
	}
	return parsed.String(), nil
}

func githubBaseQuery(cfg connectors.RuntimeConfig) url.Values {
	values := url.Values{}
	state := strings.TrimSpace(cfg.Config["state"])
	if state == "" {
		state = "all"
	}
	values.Set("state", state)
	if sort := strings.TrimSpace(cfg.Config["sort"]); sort != "" {
		values.Set("sort", sort)
	}
	if direction := strings.TrimSpace(cfg.Config["direction"]); direction != "" {
		values.Set("direction", direction)
	}
	return values
}

func githubPositiveInt(raw string, fallback, minValue, maxValue int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("github config %s must be an integer: %w", name, err)
	}
	if value < minValue || value > maxValue {
		return 0, fmt.Errorf("github config %s must be between %d and %d", name, minValue, maxValue)
	}
	return value, nil
}

func githubMaxPages(raw string) (int, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return githubDefaultMaxPages, nil
	}
	if value == "all" || value == "unlimited" {
		return 0, nil
	}
	maxPages, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("github config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if maxPages < 0 {
		return 0, errors.New("github config max_pages must be 0 for unlimited or a positive integer")
	}
	return maxPages, nil
}

func githubToken(cfg connectors.RuntimeConfig) string {
	return strings.TrimSpace(firstNonEmptyString(
		cfg.Secrets["token"],
		cfg.Secrets["personalAccessToken"],
		cfg.Secrets["accessToken"],
		cfg.Secrets["oauthToken"],
		cfg.Secrets["installationToken"],
		cfg.Secrets["githubToken"],
		cfg.Secrets["GITHUB_TOKEN"],
	))
}

func githubIssueRecord(repository string, item map[string]any) connectors.Record {
	record := githubCommonRecord(repository, item)
	record["is_pull_request"] = false
	record["state_reason"] = item["state_reason"]
	record["labels_count"] = lenAnySlice(item["labels"])
	record["assignees_count"] = lenAnySlice(item["assignees"])
	return record
}

func githubPullRequestRecord(repository string, item map[string]any) connectors.Record {
	record := githubCommonRecord(repository, item)
	record["merged_at"] = item["merged_at"]
	record["draft"] = item["draft"]
	record["merge_commit_sha"] = item["merge_commit_sha"]
	record["base_ref"] = nestedString(item, "base", "ref")
	record["base_sha"] = nestedString(item, "base", "sha")
	record["head_ref"] = nestedString(item, "head", "ref")
	record["head_sha"] = nestedString(item, "head", "sha")
	return record
}

func githubCommonRecord(repository string, item map[string]any) connectors.Record {
	return connectors.Record{
		"repository":         repository,
		"id":                 item["id"],
		"node_id":            item["node_id"],
		"number":             item["number"],
		"state":              item["state"],
		"title":              item["title"],
		"body":               item["body"],
		"html_url":           item["html_url"],
		"url":                item["url"],
		"user_login":         nestedString(item, "user", "login"),
		"user_id":            nestedValue(item, "user", "id"),
		"author_association": item["author_association"],
		"comments":           item["comments"],
		"locked":             item["locked"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
		"closed_at":          item["closed_at"],
	}
}

func nestedString(item map[string]any, key, nestedKey string) string {
	if value, ok := nestedValue(item, key, nestedKey).(string); ok {
		return value
	}
	return ""
}

func nestedValue(item map[string]any, key, nestedKey string) any {
	nested, ok := item[key].(map[string]any)
	if !ok {
		return nil
	}
	return nested[nestedKey]
}

func lenAnySlice(value any) int {
	items, ok := value.([]any)
	if !ok {
		return 0
	}
	return len(items)
}

func githubIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "repository", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "node_id", Type: "string"},
		{Name: "number", Type: "integer"},
		{Name: "state", Type: "string"},
		{Name: "state_reason", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "body", Type: "string"},
		{Name: "html_url", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "user_login", Type: "string"},
		{Name: "user_id", Type: "integer"},
		{Name: "author_association", Type: "string"},
		{Name: "comments", Type: "integer"},
		{Name: "locked", Type: "boolean"},
		{Name: "labels_count", Type: "integer"},
		{Name: "assignees_count", Type: "integer"},
		{Name: "is_pull_request", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "closed_at", Type: "timestamp"},
	}
}

func githubPullRequestFields() []connectors.Field {
	fields := githubIssueFields()
	out := make([]connectors.Field, 0, len(fields)+5)
	for _, field := range fields {
		if field.Name == "state_reason" || field.Name == "labels_count" || field.Name == "assignees_count" || field.Name == "is_pull_request" {
			continue
		}
		out = append(out, field)
	}
	out = append(out,
		connectors.Field{Name: "merged_at", Type: "timestamp"},
		connectors.Field{Name: "draft", Type: "boolean"},
		connectors.Field{Name: "merge_commit_sha", Type: "string"},
		connectors.Field{Name: "base_ref", Type: "string"},
		connectors.Field{Name: "base_sha", Type: "string"},
		connectors.Field{Name: "head_ref", Type: "string"},
		connectors.Field{Name: "head_sha", Type: "string"},
	)
	return out
}

func githubWriteActionSpecs() []connectors.WriteActionSpec {
	return []connectors.WriteActionSpec{
		{
			Name:           "create_issue",
			Description:    "Create a repository issue.",
			RequiredFields: []string{"title"},
			OptionalFields: []string{"body", "labels", "assignees", "milestone", "type"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/issues",
			Risk:           "creates user-visible GitHub issue and may notify watchers",
		},
		{
			Name:           "update_issue",
			Description:    "Edit issue title, body, state, labels, assignees, milestone, or type.",
			RequiredFields: []string{"issue_number or number"},
			OptionalFields: []string{"title", "body", "state", "state_reason", "labels", "assignees", "milestone", "type"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/issues/{issue_number}",
			Risk:           "mutates existing GitHub issue or pull request issue metadata",
		},
		{
			Name:           "comment_issue",
			Description:    "Create a comment on an issue or pull request.",
			RequiredFields: []string{"issue_number, pull_number, or number", "body"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/issues/{issue_number}/comments",
			Risk:           "creates user-visible comment and may notify participants",
		},
		{
			Name:           "close_issue",
			Description:    "Close an issue, optionally with a comment and state reason.",
			RequiredFields: []string{"issue_number or number"},
			OptionalFields: []string{"comment", "state_reason"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/issues/{issue_number}",
			Risk:           "closes existing GitHub issue",
		},
		{
			Name:           "create_pull_request",
			Description:    "Create a pull request and optionally add labels, assignees, milestone, or reviewers.",
			RequiredFields: []string{"title", "head", "base"},
			OptionalFields: []string{"body", "draft", "maintainer_can_modify", "issue", "labels", "assignees", "milestone", "reviewers", "team_reviewers"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/pulls",
			Risk:           "creates user-visible pull request and may notify watchers/reviewers",
		},
		{
			Name:           "update_pull_request",
			Description:    "Edit pull request fields and optionally add issue metadata or reviewers.",
			RequiredFields: []string{"pull_number or number"},
			OptionalFields: []string{"title", "body", "state", "base", "maintainer_can_modify", "labels", "assignees", "milestone", "reviewers", "team_reviewers"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/pulls/{pull_number}",
			Risk:           "mutates existing GitHub pull request",
		},
		{
			Name:           "close_pull_request",
			Description:    "Close a pull request, optionally with a comment.",
			RequiredFields: []string{"pull_number or number"},
			OptionalFields: []string{"comment"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/pulls/{pull_number}",
			Risk:           "closes existing GitHub pull request",
		},
		{
			Name:           "request_reviewers",
			Description:    "Request user or team reviewers for a pull request.",
			RequiredFields: []string{"pull_number or number", "reviewers or team_reviewers"},
			OptionalFields: []string{"reviewers", "team_reviewers"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers",
			Risk:           "notifies requested GitHub reviewers",
		},
		{
			Name:           "merge_pull_request",
			Description:    "Merge a pull request with optional commit title, message, SHA guard, and method.",
			RequiredFields: []string{"pull_number or number"},
			OptionalFields: []string{"commit_title", "commit_message", "sha", "merge_method"},
			Method:         "PUT",
			Path:           "/repos/{owner}/{repo}/pulls/{pull_number}/merge",
			Risk:           "irreversibly changes repository history unless branch protection blocks merge",
		},
		{
			Name:           "create_label",
			Description:    "Create a repository label.",
			RequiredFields: []string{"name", "color"},
			OptionalFields: []string{"description"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/labels",
			Risk:           "changes repository taxonomy used by issues and pull requests",
		},
		{
			Name:           "update_label",
			Description:    "Update a repository label name, color, or description.",
			RequiredFields: []string{"name"},
			OptionalFields: []string{"new_name", "color", "description"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/labels/{name}",
			Risk:           "renames or changes labels already used by issues and pull requests",
		},
		{
			Name:           "delete_label",
			Description:    "Delete a repository label.",
			RequiredFields: []string{"name"},
			Method:         "DELETE",
			Path:           "/repos/{owner}/{repo}/labels/{name}",
			Risk:           "removes a label from the repository and existing issue metadata",
		},
		{
			Name:           "create_milestone",
			Description:    "Create a repository milestone.",
			RequiredFields: []string{"title"},
			OptionalFields: []string{"state", "description", "due_on"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/milestones",
			Risk:           "creates planning metadata visible to repository collaborators",
		},
		{
			Name:           "update_milestone",
			Description:    "Update milestone title, state, description, or due date.",
			RequiredFields: []string{"milestone_number or number"},
			OptionalFields: []string{"title", "state", "description", "due_on"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/milestones/{milestone_number}",
			Risk:           "changes planning metadata used by issues and pull requests",
		},
		{
			Name:           "delete_milestone",
			Description:    "Delete a repository milestone.",
			RequiredFields: []string{"milestone_number or number"},
			Method:         "DELETE",
			Path:           "/repos/{owner}/{repo}/milestones/{milestone_number}",
			Risk:           "removes repository planning metadata from GitHub",
		},
		{
			Name:           "create_release",
			Description:    "Create a repository release for a tag.",
			RequiredFields: []string{"tag_name"},
			OptionalFields: []string{"target_commitish", "name", "body", "draft", "prerelease", "generate_release_notes", "make_latest"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/releases",
			Risk:           "publishes release metadata and may notify repository watchers",
		},
		{
			Name:           "update_release",
			Description:    "Update release metadata.",
			RequiredFields: []string{"release_id or id"},
			OptionalFields: []string{"tag_name", "target_commitish", "name", "body", "draft", "prerelease", "generate_release_notes", "make_latest"},
			Method:         "PATCH",
			Path:           "/repos/{owner}/{repo}/releases/{release_id}",
			Risk:           "changes published release metadata",
		},
		{
			Name:           "delete_release",
			Description:    "Delete a repository release.",
			RequiredFields: []string{"release_id or id"},
			Method:         "DELETE",
			Path:           "/repos/{owner}/{repo}/releases/{release_id}",
			Risk:           "removes release metadata from GitHub; tags are not deleted by this action",
		},
		{
			Name:           "dispatch_workflow",
			Description:    "Trigger a GitHub Actions workflow dispatch event.",
			RequiredFields: []string{"workflow_id", "ref"},
			OptionalFields: []string{"inputs"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches",
			Risk:           "starts CI/CD automation that may deploy, publish, or mutate external systems",
		},
		{
			Name:           "rerun_workflow_run",
			Description:    "Rerun a GitHub Actions workflow run.",
			RequiredFields: []string{"run_id, workflow_run_id, or id"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/actions/runs/{run_id}/rerun",
			Risk:           "reruns CI/CD automation and consumes workflow minutes",
		},
		{
			Name:           "cancel_workflow_run",
			Description:    "Cancel a GitHub Actions workflow run.",
			RequiredFields: []string{"run_id, workflow_run_id, or id"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/actions/runs/{run_id}/cancel",
			Risk:           "interrupts in-flight CI/CD automation",
		},
		{
			Name:           "delete_workflow_run",
			Description:    "Delete a GitHub Actions workflow run record.",
			RequiredFields: []string{"run_id, workflow_run_id, or id"},
			Method:         "DELETE",
			Path:           "/repos/{owner}/{repo}/actions/runs/{run_id}",
			Risk:           "removes workflow run history from GitHub",
		},
		{
			Name:           "create_pull_request_review",
			Description:    "Create a pull request review with optional review comments.",
			RequiredFields: []string{"pull_number or number"},
			OptionalFields: []string{"event", "body", "commit_id", "comments"},
			Method:         "POST",
			Path:           "/repos/{owner}/{repo}/pulls/{pull_number}/reviews",
			Risk:           "submits reviewer feedback and may approve or request changes on a pull request",
		},
		{
			Name:           "create_or_update_file",
			Description:    "Create or update repository file contents.",
			RequiredFields: []string{"path", "message", "content or content_base64"},
			OptionalFields: []string{"sha", "branch", "committer", "author"},
			Method:         "PUT",
			Path:           "/repos/{owner}/{repo}/contents/{path}",
			Risk:           "writes a commit to the repository and may trigger CI/CD",
		},
		{
			Name:           "delete_file",
			Description:    "Delete a repository file through the contents API.",
			RequiredFields: []string{"path", "message", "sha"},
			OptionalFields: []string{"branch", "committer", "author"},
			Method:         "DELETE",
			Path:           "/repos/{owner}/{repo}/contents/{path}",
			Risk:           "writes a commit that removes a file from the repository",
		},
	}
}
