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

var errRepeatedPaginationToken = errors.New("repeated pagination next token")

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
		Description:     "Reads and safely operates AWS CloudTrail trails, event data stores, channels, dashboards, and Lake queries through fixed AWS JSON-RPC streams, typed direct-read commands, and typed reverse-ETL write/admin actions. Only StartQuery, CreateDashboard, and UpdateDashboard stay blocked: each requires an unrestricted CloudTrail Lake SQL QueryStatement, which this project disables for every connector by policy. Breaking change: the earlier LookupEvents-backed management_events, read_only_events, write_only_events, and console_logins streams are removed and have no replacement here; use the events lookup direct-read command for CloudTrail event lookups instead. See the connector docs migration note.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: true, Query: false},
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
	streams, err := streams(ctx)
	if err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
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
	body, err := buildActionBodyFromStrings(action, req.Query, true)
	if err == nil {
		return c.emitActionBody(ctx, req.Config, action, body, emit)
	}
	if len(req.Query) > 0 || !cloudTrailCanDeriveActionBody(action) {
		return err
	}
	bodies, err := c.derivedActionBodies(ctx, req.Config, action)
	if err != nil {
		return err
	}
	for _, body := range bodies {
		if err := c.emitActionBody(ctx, req.Config, action, body, emit); err != nil {
			if shouldSkipDerivedActionError(action, err) {
				continue
			}
			return err
		}
	}
	return nil
}

func (c Connector) emitActionBody(ctx context.Context, cfg connectors.RuntimeConfig, action string, rawBody map[string]any, emit func(connectors.Record) error) error {
	r, err := c.requester(cfg, action)
	if err != nil {
		return err
	}
	body := copyActionBody(rawBody)
	if supportsField(action, "MaxResults") {
		maxItems, err := pageSizeForAction(action, cfg)
		if err != nil {
			return err
		}
		if _, ok := body["MaxResults"]; !ok {
			body["MaxResults"] = maxItems
		}
	}
	maxPageLimit, err := maxPages(cfg)
	if err != nil {
		return err
	}
	seenTokens := map[string]bool{}
	if token, ok := stringAt(body, "NextToken"); ok {
		if token = strings.TrimSpace(token); token != "" {
			seenTokens[token] = true
		}
	}
	for page := 0; maxPageLimit == 0 || page < maxPageLimit; page++ {
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
		if err := emitActionRecords(action, decoded, body, emit); err != nil {
			return err
		}
		next, _ := stringAt(decoded, "NextToken")
		trimmedNext := strings.TrimSpace(next)
		if trimmedNext == "" {
			return nil
		}
		if maxPageLimit > 0 && page+1 >= maxPageLimit {
			return nil
		}
		if seenTokens[trimmedNext] {
			return fmt.Errorf("read aws-cloudtrail %s: %w %q", action, errRepeatedPaginationToken, trimmedNext)
		}
		seenTokens[trimmedNext] = true
		body["NextToken"] = next
	}
	return nil
}

func copyActionBody(body map[string]any) map[string]any {
	out := make(map[string]any, len(body))
	for key, value := range body {
		out[key] = value
	}
	return out
}

func cloudTrailCanDeriveActionBody(action string) bool {
	switch action {
	case "GetChannel", "GetDashboard", "GetEventConfiguration", "GetEventDataStore", "GetEventSelectors", "GetImport", "GetInsightSelectors", "GetResourcePolicy", "GetTrail", "GetTrailStatus", "ListImportFailures", "ListTags":
		return true
	default:
		return false
	}
}

var cloudTrailTrailOrEventDataStoreActions = map[string][]string{
	"GetEventConfiguration": nil,
	"GetInsightSelectors":   {"insightnotenabledexception", "insight not enabled", "insights not enabled"},
}

var cloudTrailTrailOrEventDataStoreSkipMarkers = []string{
	"trailnotfoundexception",
	"eventdatastorenotfoundexception",
	"inactiveeventdatastoreexception",
	"invalideventdatastorecategoryexception",
	"unsupportedoperationexception",
}

func (c Connector) derivedActionBodies(ctx context.Context, cfg connectors.RuntimeConfig, action string) ([]map[string]any, error) {
	if _, ok := cloudTrailTrailOrEventDataStoreActions[action]; ok {
		return c.derivedTrailOrEventDataStoreBodies(ctx, cfg, action)
	}
	switch action {
	case "GetChannel":
		return c.derivedBodiesFromDiscovery(ctx, cfg, action, "ListChannels", "Channel", []string{"ChannelArn", "Channel", "Arn"})
	case "GetDashboard":
		return c.derivedBodiesFromDiscovery(ctx, cfg, action, "ListDashboards", "DashboardId", []string{"DashboardId", "DashboardArn", "Arn"})
	case "GetEventDataStore":
		return c.derivedBodiesFromDiscovery(ctx, cfg, action, "ListEventDataStores", "EventDataStore", []string{"EventDataStoreArn", "EventDataStore", "Arn"})
	case "GetEventSelectors":
		return c.derivedBodiesFromDiscovery(ctx, cfg, action, "DescribeTrails", "TrailName", []string{"TrailARN", "TrailArn", "Name", "TrailName"})
	case "GetImport", "ListImportFailures":
		return c.derivedBodiesFromDiscovery(ctx, cfg, action, "ListImports", "ImportId", []string{"ImportId"})
	case "GetResourcePolicy":
		values, err := c.discoveredResourceARNs(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return validatedDerivedBodies(action, bodiesForValues("ResourceArn", values))
	case "GetTrail", "GetTrailStatus":
		return c.derivedBodiesFromDiscovery(ctx, cfg, action, "DescribeTrails", "Name", []string{"TrailARN", "TrailArn", "Name", "TrailName"})
	case "ListTags":
		values, err := c.discoveredTagResourceIDs(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return validatedDerivedBodies(action, bodiesForStringChunks("ResourceIdList", values, 20))
	default:
		return nil, fmt.Errorf("aws-cloudtrail %s cannot derive required request fields", action)
	}
}

func (c Connector) derivedTrailOrEventDataStoreBodies(ctx context.Context, cfg connectors.RuntimeConfig, action string) ([]map[string]any, error) {
	trailValues, err := c.discoveryValues(ctx, cfg, "DescribeTrails", "TrailARN", "TrailArn", "Name", "TrailName")
	if err != nil {
		return nil, err
	}
	eventDataStoreValues, err := c.discoveryValues(ctx, cfg, "ListEventDataStores", "EventDataStoreArn", "EventDataStore", "Arn")
	if err != nil {
		return nil, err
	}
	bodies := bodiesForValues("TrailName", trailValues)
	bodies = append(bodies, bodiesForValues("EventDataStore", eventDataStoreValues)...)
	return validatedDerivedBodies(action, bodies)
}

func (c Connector) derivedBodiesFromDiscovery(ctx context.Context, cfg connectors.RuntimeConfig, action, discoveryAction, requestField string, keys []string) ([]map[string]any, error) {
	values, err := c.discoveryValues(ctx, cfg, discoveryAction, keys...)
	if err != nil {
		return nil, err
	}
	return validatedDerivedBodies(action, bodiesForValues(requestField, values))
}

func (c Connector) discoveryValues(ctx context.Context, cfg connectors.RuntimeConfig, action string, keys ...string) ([]string, error) {
	var values []string
	seen := map[string]bool{}
	err := c.emitActionBody(ctx, cfg, action, nil, func(record connectors.Record) error {
		for _, key := range keys {
			value, ok := stringAt(map[string]any(record), key)
			value = strings.TrimSpace(value)
			if !ok || value == "" || seen[value] {
				continue
			}
			seen[value] = true
			values = append(values, value)
			break
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (c Connector) discoveredResourceARNs(ctx context.Context, cfg connectors.RuntimeConfig) ([]string, error) {
	discoveries := []struct {
		action string
		keys   []string
	}{
		{action: "ListChannels", keys: []string{"ChannelArn", "Arn"}},
		{action: "ListDashboards", keys: []string{"DashboardArn", "DashboardId", "Arn"}},
		{action: "ListEventDataStores", keys: []string{"EventDataStoreArn", "EventDataStore", "Arn"}},
	}
	return c.discoveryValuesFromMany(ctx, cfg, discoveries)
}

func (c Connector) discoveredTagResourceIDs(ctx context.Context, cfg connectors.RuntimeConfig) ([]string, error) {
	discoveries := []struct {
		action string
		keys   []string
	}{
		{action: "DescribeTrails", keys: []string{"TrailARN", "TrailArn", "Arn"}},
		{action: "ListChannels", keys: []string{"ChannelArn", "Arn"}},
		{action: "ListDashboards", keys: []string{"DashboardArn", "DashboardId", "Arn"}},
		{action: "ListEventDataStores", keys: []string{"EventDataStoreArn", "EventDataStore", "Arn"}},
	}
	return c.discoveryValuesFromMany(ctx, cfg, discoveries)
}

func (c Connector) discoveryValuesFromMany(ctx context.Context, cfg connectors.RuntimeConfig, discoveries []struct {
	action string
	keys   []string
}) ([]string, error) {
	var values []string
	seen := map[string]bool{}
	for _, discovery := range discoveries {
		items, err := c.discoveryValues(ctx, cfg, discovery.action, discovery.keys...)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if seen[item] {
				continue
			}
			seen[item] = true
			values = append(values, item)
		}
	}
	return values, nil
}

func bodiesForValues(field string, values []string) []map[string]any {
	bodies := make([]map[string]any, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		bodies = append(bodies, map[string]any{field: value})
	}
	return bodies
}

func bodiesForStringChunks(field string, values []string, size int) []map[string]any {
	if size <= 0 {
		size = len(values)
	}
	var bodies []map[string]any
	for start := 0; start < len(values); start += size {
		end := start + size
		if end > len(values) {
			end = len(values)
		}
		chunk := append([]string(nil), values[start:end]...)
		if len(chunk) > 0 {
			bodies = append(bodies, map[string]any{field: chunk})
		}
	}
	return bodies
}

func validatedDerivedBodies(action string, bodies []map[string]any) ([]map[string]any, error) {
	for i, body := range bodies {
		if _, err := buildActionBody(action, body, true); err != nil {
			return nil, fmt.Errorf("derived request body %d: %w", i, err)
		}
	}
	return bodies, nil
}

func shouldSkipDerivedActionError(action string, err error) bool {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	body := strings.ToLower(httpErr.Body)
	if actionMarkers, ok := cloudTrailTrailOrEventDataStoreActions[action]; ok {
		if httpErr.Status == http.StatusNotFound {
			return true
		}
		if httpErr.Status != http.StatusBadRequest {
			return false
		}
		return containsAny(body, cloudTrailTrailOrEventDataStoreSkipMarkers...) || containsAny(body, actionMarkers...)
	}
	switch action {
	case "GetResourcePolicy":
		return httpErr.Status == http.StatusNotFound || containsAny(body, "notfound", "not found")
	default:
		return false
	}
}

func containsAny(value string, markers ...string) bool {
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

// OperationDirectRead executes only the declared, typed CloudTrail lookup
// operations. The promoted-native wrapper forwards this optional capability
// into the command runner.
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
		maxItems, err := pageSizeForAction(action, req.Config)
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
	limit, err := operationReadMaxBytes(req.Operation, req.MaxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
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

func pageSizeForAction(action string, cfg connectors.RuntimeConfig) (int, error) {
	value, err := pageSize(cfg)
	if err != nil {
		return 0, err
	}
	constraints := cloudTrailActionFieldConstraints[action]["MaxResults"]
	if constraints.maximum > 0 && value > constraints.maximum {
		return constraints.maximum, nil
	}
	return value, nil
}

func operationReadMaxBytes(operation string, requested int) (int, error) {
	for _, candidate := range cloudTrailBundle().Operations {
		if candidate.ID != operation {
			continue
		}
		if candidate.REST == nil || candidate.REST.MaxBytes <= 0 {
			return 0, fmt.Errorf("aws-cloudtrail operation %q has no positive response limit", operation)
		}
		if requested <= 0 || requested > candidate.REST.MaxBytes {
			return candidate.REST.MaxBytes, nil
		}
		return requested, nil
	}
	return 0, fmt.Errorf("aws-cloudtrail operation %q is not declared", operation)
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" || raw == "synthetic-conformance-value" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aws-cloudtrail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("aws-cloudtrail config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
