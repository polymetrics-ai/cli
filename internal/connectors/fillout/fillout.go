// Package fillout implements the native pm Fillout connector. It is a declarative
// HTTP per-system connector built on the connsdk toolkit (Requester + Bearer auth
// + RecordsAt extraction) following the stripe reference shape.
//
// Fillout is a form builder; its REST API (https://www.fillout.com/help/fillout-rest-api)
// exposes the account's forms, their question definitions, and the submissions
// (responses) for each form. The connector is read-only: creating or deleting
// submissions is not a safe reverse-ETL operation, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package fillout

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
	filloutDefaultBaseURL  = "https://api.fillout.com/v1/api"
	filloutDefaultPageSize = 50
	filloutMaxPageSize     = 150
	filloutUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("fillout", New)
}

// New returns the Fillout connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Fillout connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "fillout" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "fillout",
		DisplayName:     "Fillout",
		IntegrationType: "api",
		Description:     "Reads Fillout forms, their question definitions, and form submissions (responses) through the Fillout REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Fillout. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := filloutBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(filloutSecret(cfg)) == "" {
		return errors.New("fillout connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing forms is a bounded, read-only call that confirms auth + connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "forms", nil, nil, nil); err != nil {
		return fmt.Errorf("check fillout: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: filloutStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a submissions stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "forms"
	}
	switch stream {
	case "forms", "questions", "submissions":
	default:
		return fmt.Errorf("fillout stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch stream {
	case "forms":
		return c.readForms(ctx, r, emit)
	case "questions":
		return c.readQuestions(ctx, r, req, emit)
	default:
		return c.readSubmissions(ctx, r, req, emit)
	}
}

// Write is unsupported: Fillout submissions are inbound user data, not a safe
// reverse-ETL write target, so the connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readForms emits one record per form. GET /forms returns a top-level JSON array.
func (c Connector) readForms(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	forms, err := c.listForms(ctx, r)
	if err != nil {
		return err
	}
	for _, form := range forms {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(filloutFormRecord(form)); err != nil {
			return err
		}
	}
	return nil
}

// readQuestions fans out across every form (or the configured form_id), emitting
// each question of each form's metadata (GET /forms/{id}).
func (c Connector) readQuestions(ctx context.Context, r *connsdk.Requester, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	formIDs, err := c.resolveFormIDs(ctx, r, req.Config)
	if err != nil {
		return err
	}
	for _, formID := range formIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, "forms/"+url.PathEscape(formID), nil, nil)
		if err != nil {
			return fmt.Errorf("read fillout form %s: %w", formID, err)
		}
		questions, err := connsdk.RecordsAt(resp.Body, "questions")
		if err != nil {
			return fmt.Errorf("decode fillout form %s questions: %w", formID, err)
		}
		for _, q := range questions {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(filloutQuestionRecord(formID, q)); err != nil {
				return err
			}
		}
	}
	return nil
}

// readSubmissions fans out across every form (or the configured form_id),
// paginating each form's submissions via offset/limit until a short page.
func (c Connector) readSubmissions(ctx context.Context, r *connsdk.Requester, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	formIDs, err := c.resolveFormIDs(ctx, r, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := filloutPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := filloutMaxPages(req.Config)
	if err != nil {
		return err
	}
	afterDate := incrementalLowerBound(req)

	for _, formID := range formIDs {
		if err := c.harvestSubmissions(ctx, r, formID, pageSize, maxPages, afterDate, emit); err != nil {
			return err
		}
	}
	return nil
}

// harvestSubmissions drives Fillout's offset/limit pagination for one form. The
// submissions endpoint returns {responses:[...], totalResponses, pageCount}; a
// page shorter than the limit signals the end.
func (c Connector) harvestSubmissions(ctx context.Context, r *connsdk.Requester, formID string, pageSize, maxPages int, afterDate string, emit func(connectors.Record) error) error {
	path := "forms/" + url.PathEscape(formID) + "/submissions"
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		if afterDate != "" {
			query.Set("afterDate", afterDate)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read fillout submissions for %s: %w", formID, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "responses")
		if err != nil {
			return fmt.Errorf("decode fillout submissions for %s: %w", formID, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(filloutSubmissionRecord(formID, item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// resolveFormIDs returns the form ids to read submissions/questions for: either
// the explicitly configured form_id (comma-separated), or every form discovered
// via GET /forms.
func (c Connector) resolveFormIDs(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig) ([]string, error) {
	if raw := strings.TrimSpace(cfg.Config["form_id"]); raw != "" {
		var ids []string
		for _, part := range strings.Split(raw, ",") {
			if id := strings.TrimSpace(part); id != "" {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			return ids, nil
		}
	}
	forms, err := c.listForms(ctx, r)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(forms))
	for _, form := range forms {
		if id, ok := firstString(form, "formId", "id").(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// listForms fetches the top-level forms array from GET /forms.
func (c Connector) listForms(ctx context.Context, r *connsdk.Requester) ([]map[string]any, error) {
	resp, err := r.Do(ctx, http.MethodGet, "forms", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("read fillout forms: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return nil, fmt.Errorf("decode fillout forms: %w", err)
	}
	out := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		out = append(out, map[string]any(rec))
	}
	return out, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise fillout credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	const formID = "fillout_form_fixture_1"
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var record connectors.Record
		switch stream {
		case "forms":
			record = filloutFormRecord(map[string]any{
				"formId": fmt.Sprintf("fillout_form_fixture_%d", i),
				"name":   fmt.Sprintf("Fixture Form %d", i),
			})
		case "questions":
			record = filloutQuestionRecord(formID, map[string]any{
				"id":   fmt.Sprintf("q_fixture_%d", i),
				"name": fmt.Sprintf("Fixture Question %d", i),
				"type": "ShortAnswer",
			})
		default: // submissions
			record = filloutSubmissionRecord(formID, map[string]any{
				"submissionId":   fmt.Sprintf("sub_fixture_%d", i),
				"submissionTime": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
				"lastUpdatedAt":  fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
				"questions": []any{map[string]any{
					"id": "q_fixture_1", "name": "Fixture Question 1", "type": "ShortAnswer", "value": fmt.Sprintf("answer %d", i),
				}},
			})
		}
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := filloutBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := filloutSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("fillout connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: filloutUserAgent,
	}, nil
}

// incrementalLowerBound returns the afterDate lower bound for submissions,
// derived from the incremental cursor (if any) or else the start_date config.
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func filloutSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// filloutBaseURL resolves and validates the base URL. The default is
// api.fillout.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func filloutBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return filloutDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("fillout config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("fillout config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("fillout config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func filloutPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return filloutDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fillout config page_size must be an integer: %w", err)
	}
	if value < 1 || value > filloutMaxPageSize {
		return 0, fmt.Errorf("fillout config page_size must be between 1 and %d", filloutMaxPageSize)
	}
	return value, nil
}

func filloutMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fillout config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("fillout config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
