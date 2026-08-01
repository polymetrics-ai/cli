// Package awscloudtrail implements the native pm AWS CloudTrail connector.
//
// CloudTrail uses a JSON-RPC style AWS API: every operation is a signed POST to
// / with an operation-specific X-Amz-Target header. This package keeps that
// action selection connector-local and closed over the operation ledger in the
// aws-cloudtrail bundle; callers never supply raw AWS action names, paths,
// headers, or bodies.
package awscloudtrail

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName = "aws-cloudtrail"

	cloudTrailService      = "cloudtrail"
	cloudTrailTargetPrefix = "com.amazonaws.cloudtrail.v20131101.CloudTrail_20131101."
	cloudTrailJSONType     = "application/x-amz-json-1.1"

	defaultRegion   = "us-east-1"
	defaultMaxItems = 50
	maxMaxItems     = 1000
	userAgent       = "polymetrics-go-cli"
)

// New returns the AWS CloudTrail connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm AWS CloudTrail connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "AWS CloudTrail",
		IntegrationType: "api",
		Description:     "Reads AWS CloudTrail configuration lists through fixed AWS JSON-RPC streams that need no per-call resource identifiers. Provider query/direct-read, parameterized read, and write/admin actions remain planned until safe shared forwarding exists.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false, Query: false},
	}
}

// Check verifies SigV4 credentials and endpoint reachability with a bounded
// read-only DescribeTrails call. In fixture mode it short-circuits without a
// network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg, "DescribeTrails")
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodPost, "/", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("check aws-cloudtrail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "describe_trails"
	}
	action, ok := cloudTrailStreamActions[stream]
	if !ok || !cloudTrailStreamPublished(stream) {
		return fmt.Errorf("aws-cloudtrail stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, action, req, emit)
	}
	return c.readAction(ctx, action, stream, req, emit)
}

func cloudTrailStreamPublished(stream string) bool {
	for _, published := range cloudTrailPublishedStreams {
		if stream == published {
			return true
		}
	}
	return false
}

func (c Connector) readAction(ctx context.Context, action, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	r, err := c.requester(req.Config, action)
	if err != nil {
		return err
	}
	body, err := buildActionBodyFromStrings(action, req.Query, true)
	if err != nil {
		return err
	}
	if supportsField(action, "MaxResults") {
		maxItems, err := pageSize(req.Config)
		if err != nil {
			return err
		}
		if _, ok := body["MaxResults"]; !ok {
			body["MaxResults"] = maxItems
		}
	}
	maxPages := maxPages(req.Config)
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodPost, "/", nil, body)
		if err != nil {
			return fmt.Errorf("read aws-cloudtrail %s: %w", action, err)
		}
		var decoded map[string]any
		if err := decodeJSON(resp.Body, &decoded); err != nil {
			return fmt.Errorf("decode aws-cloudtrail %s response: %w", action, err)
		}
		if err := emitActionRecords(action, decoded, emit); err != nil {
			return err
		}
		next, _ := stringAt(decoded, "NextToken")
		if strings.TrimSpace(next) == "" {
			return nil
		}
		body["NextToken"] = next
	}
	return nil
}

// OperationDirectRead rejects provider query/lookup operations until shared
// promoted-native forwarding exposes them safely.
func (c Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	action, ok := cloudTrailDirectOperations[req.Operation]
	if !ok {
		return connectors.DirectReadResult{}, fmt.Errorf("aws-cloudtrail operation %q is not a declared direct read", req.Operation)
	}
	body, err := buildActionBody(action, req.Body, true)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if supportsField(action, "MaxResults") {
		maxItems, err := pageSize(req.Config)
		if err != nil {
			return connectors.DirectReadResult{}, err
		}
		if _, ok := body["MaxResults"]; !ok {
			body["MaxResults"] = maxItems
		}
	}
	r, err := c.requester(req.Config, action)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	limit := directMaxBytes(req.MaxBytes)
	resp, err := r.DoLimited(ctx, http.MethodPost, "/", nil, body, limit)
	if err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read aws-cloudtrail %s: %w", action, err)
	}
	if len(resp.Body) > limit {
		return connectors.DirectReadResult{}, connectors.ErrReadLimitReached
	}
	var decoded any
	if err := decodeJSON(resp.Body, &decoded); err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read aws-cloudtrail %s response is not JSON: %w", action, err)
	}
	decoded = redactFields(decoded, directRedactFields(req.RedactFields))
	return connectors.DirectReadResult{Connector: connectorName, Method: http.MethodPost, Path: "/", Status: resp.Status, Body: decoded}, nil
}

func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	action, ok := cloudTrailWriteActions[req.Action]
	if !ok {
		return fmt.Errorf("aws-cloudtrail write action %q not found", req.Action)
	}
	for i, rec := range records {
		if _, err := buildActionBody(action, map[string]any(rec), true); err != nil {
			return fmt.Errorf("record %d: %w", i, err)
		}
	}
	return nil
}

func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WritePreview{}, err
	}
	action := cloudTrailWriteActions[req.Action]
	warnings := []string{fmt.Sprintf("%s executes AWS CloudTrail %s only after reverse-ETL approval; dry run performs no external call", req.Action, action)}
	if len(records) > 0 {
		warnings = append(warnings, fmt.Sprintf("resolved request: POST / with X-Amz-Target %s", cloudTrailTarget(action)))
	}
	return connectors.WritePreview{RecordsStaged: len(records), Action: req.Action, Warnings: warnings}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	action := cloudTrailWriteActions[req.Action]
	r, err := c.requester(req.Config, action)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	result := connectors.WriteResult{}
	for i, rec := range records {
		if err := ctx.Err(); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, err
		}
		body, err := buildActionBody(action, map[string]any(rec), true)
		if err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("record %d: %w", i, err)
		}
		if _, err := r.Do(ctx, http.MethodPost, "/", nil, body); err != nil {
			if isMissingOK(req.Action, err) {
				result.RecordsWritten++
				continue
			}
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("write aws-cloudtrail %s record %d: %w", action, i, err)
		}
		result.RecordsWritten++
	}
	return result, nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig, action string) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	keyID, secret := secrets(cfg)
	if strings.TrimSpace(keyID) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("aws-cloudtrail connector requires secrets aws_key_id and aws_secret_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      &sigV4Signer{accessKeyID: strings.TrimSpace(keyID), secretAccessKey: strings.TrimSpace(secret), region: region(cfg), service: cloudTrailService},
		UserAgent: userAgent,
		Accept:    cloudTrailJSONType,
		DefaultHeaders: map[string]string{
			"Content-Type": cloudTrailJSONType,
			"X-Amz-Target": cloudTrailTarget(action),
		},
	}, nil
}

func cloudTrailTarget(action string) string { return cloudTrailTargetPrefix + action }

func secrets(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["aws_key_id"], cfg.Secrets["aws_secret_key"]
}

func region(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if r := strings.TrimSpace(cfg.Config["aws_region_name"]); r != "" {
			return r
		}
	}
	return defaultRegion
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := ""
	if cfg.Config != nil {
		base = strings.TrimSpace(cfg.Config["base_url"])
	}
	if base == "" {
		return "https://" + cloudTrailService + "." + region(cfg) + ".amazonaws.com", nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("aws-cloudtrail config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("aws-cloudtrail config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("aws-cloudtrail config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" || raw == "synthetic-conformance-value" {
		return defaultMaxItems, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aws-cloudtrail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxMaxItems {
		return 0, fmt.Errorf("aws-cloudtrail config page_size must be between 1 and %d", maxMaxItems)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) int {
	if cfg.Config == nil {
		return 0
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" || raw == "synthetic-conformance-value" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
