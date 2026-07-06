// Package awscloudtrail implements the native pm AWS CloudTrail connector. It is
// a declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk Requester with
// CloudTrail-specific stream definitions and read logic.
//
// CloudTrail is read-only here. Its single underlying read action, LookupEvents,
// returns management events for the trailing 90 days; the connector exposes a
// handful of convenience streams that are the same LookupEvents call narrowed by
// a server-side LookupAttributes filter. Authentication is AWS Signature V4,
// implemented in sigv4.go with stdlib crypto only (no AWS SDK dependency).
package awscloudtrail

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName = "aws-cloudtrail"
	// cloudTrailService and apiVersion identify the SigV4 service and the
	// X-Amz-Target action header CloudTrail's JSON-RPC endpoint requires.
	cloudTrailService  = "cloudtrail"
	lookupEventsTarget = "com.amazonaws.cloudtrail.v20131101.CloudTrail_20131101.LookupEvents"
	cloudTrailJSONType = "application/x-amz-json-1.1"

	defaultRegion   = "us-east-1"
	defaultMaxItems = 50
	maxMaxItems     = 50
	userAgent       = "polymetrics-go-cli"

	// fixtureEventTime is the deterministic EventTime (unix seconds) used by
	// fixture-mode records (2026-01-01T00:00:00Z).
	fixtureEventTime int64 = 1767225600
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
		Description:     "Reads AWS CloudTrail management events (last 90 days) via the LookupEvents API using AWS Signature V4 authentication. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to CloudTrail.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	keyID, secret := secrets(cfg)
	if strings.TrimSpace(keyID) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("aws-cloudtrail connector requires secrets aws_key_id and aws_secret_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded LookupEvents with MaxResults=1 confirms auth + connectivity
	// without mutating anything (CloudTrail LookupEvents is read-only).
	body := map[string]any{"MaxResults": 1}
	if err := r.DoJSON(ctx, http.MethodPost, "/", nil, body, nil); err != nil {
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

// InitialState satisfies connectors.StatefulReader: a CloudTrail stream starts
// with an empty incremental cursor (full sync), which start_date can raise at
// read time.
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
		stream = "management_events"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("aws-cloudtrail stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, spec, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxItems, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	startTime, err := startTimeBound(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, spec, maxItems, startTime, req.Config, emit)
}

// harvest drives CloudTrail's NextToken pagination. LookupEvents returns
// {"Events":[...], "NextToken":"..."}; the next page resends the same body with
// NextToken set. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, spec streamSpec, maxItems int, startTime *time.Time, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	lookupAttrs := lookupAttributes(spec, cfg)

	nextToken := ""
	maxPages := maxPages(cfg)
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := map[string]any{"MaxResults": maxItems}
		if len(lookupAttrs) > 0 {
			body["LookupAttributes"] = lookupAttrs
		}
		if startTime != nil {
			body["StartTime"] = startTime.Unix()
		}
		if nextToken != "" {
			body["NextToken"] = nextToken
		}

		resp, err := r.Do(ctx, http.MethodPost, "/", nil, body)
		if err != nil {
			return fmt.Errorf("read aws-cloudtrail LookupEvents: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "Events")
		if err != nil {
			return fmt.Errorf("decode aws-cloudtrail Events page: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		nextToken, err = connsdk.StringAt(resp.Body, "NextToken")
		if err != nil {
			return fmt.Errorf("decode aws-cloudtrail NextToken: %w", err)
		}
		if strings.TrimSpace(nextToken) == "" {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, spec streamSpec, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"EventId":         fmt.Sprintf("%s_fixture_%d", stream, i),
			"EventName":       "ConsoleLogin",
			"EventSource":     "signin.amazonaws.com",
			"EventTime":       fixtureEventTime + int64(i),
			"Username":        fmt.Sprintf("fixture-user-%d", i),
			"AccessKeyId":     "AKIAFIXTURE000000000",
			"ReadOnly":        "true",
			"Resources":       []any{},
			"CloudTrailEvent": `{"eventVersion":"1.08","fixture":true}`,
		}
		record := spec.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with SigV4 auth, the resolved base
// URL, the CloudTrail JSON-RPC headers, and the LookupEvents target. The secret
// only ever flows into the signer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	keyID, secret := secrets(cfg)
	if strings.TrimSpace(keyID) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("aws-cloudtrail connector requires secrets aws_key_id and aws_secret_key")
	}
	signer := &sigV4Signer{
		accessKeyID:     strings.TrimSpace(keyID),
		secretAccessKey: strings.TrimSpace(secret),
		region:          region(cfg),
		service:         cloudTrailService,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      signer,
		UserAgent: userAgent,
		Accept:    cloudTrailJSONType,
		DefaultHeaders: map[string]string{
			// applyHeaders sets Content-Type before DefaultHeaders, so these win.
			"Content-Type": cloudTrailJSONType,
			"X-Amz-Target": lookupEventsTarget,
		},
	}, nil
}

// lookupAttributes builds the LookupAttributes filter list from the stream spec
// and an optional operator-supplied lookup_attributes_filter config override.
func lookupAttributes(spec streamSpec, cfg connectors.RuntimeConfig) []map[string]string {
	key, value := spec.filterKey, spec.filterValue
	if k := strings.TrimSpace(cfg.Config["lookup_attribute_key"]); k != "" {
		key = k
		value = strings.TrimSpace(cfg.Config["lookup_attribute_value"])
	}
	if key == "" {
		return nil
	}
	return []map[string]string{{"AttributeKey": key, "AttributeValue": value}}
}

func secrets(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["aws_key_id"], cfg.Secrets["aws_secret_key"]
}

func region(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return defaultRegion
	}
	if r := strings.TrimSpace(cfg.Config["aws_region_name"]); r != "" {
		return r
	}
	return defaultRegion
}

// baseURL resolves and validates the CloudTrail endpoint. The default is derived
// from the region (cloudtrail.<region>.amazonaws.com); any override must be an
// absolute https (or http for local test servers) URL with a host to bound SSRF
// risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
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
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
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
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

// startTimeBound returns the StartTime lower bound for LookupEvents, derived from
// the incremental cursor (unix seconds) if present, else the start_date config
// (YYYY-MM-DD per the CloudTrail spec). A nil result means no lower bound.
func startTimeBound(req connectors.ReadRequest) (*time.Time, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		secs, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("aws-cloudtrail cursor must be unix seconds: %w", err)
		}
		t := time.Unix(secs, 0).UTC()
		return &t, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		// Also accept a full RFC3339 timestamp.
		t, err = time.Parse(time.RFC3339, startDate)
		if err != nil {
			return nil, fmt.Errorf("aws-cloudtrail config start_date must be YYYY-MM-DD: %w", err)
		}
	}
	t = t.UTC()
	return &t, nil
}

// Write is unsupported: AWS CloudTrail is a read-only source. The method exists
// only to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
