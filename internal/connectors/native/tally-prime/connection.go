package tallyprime

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

const (
	defaultGatewayURL = "http://localhost:9000"
	// jsonSysName/xmlSysName are TallyPrime's SVEXPORTFORMAT sysname values.
	// UTF8JSON is TallyPrime 7.0+'s native JSON export mode (preferred); XML
	// is the universally-supported fallback documented in docs.md's Streams
	// notes for pre-7.0 releases.
	jsonSysName = "$$SysName:UTF8JSON"
	xmlSysName  = "$$SysName:XML"
)

// tallyConfig is the validated connection configuration for one request.
// TallyPrime's Gateway Server carries no credential/token of its own (see
// docs.md's Auth setup) — there are no secrets to resolve here, unlike
// postgres's password or amazon-sqs's access/secret key pair.
type tallyConfig struct {
	gatewayURL     string
	company        string
	envelopeFormat string // "json" or "xml"
	fromDate       string
	toDate         string
	timeout        time.Duration
}

// resolveConfig validates config into a tallyConfig. gateway_url defaults to
// http://localhost:9000 (TallyPrime's default Gateway Server port); company
// is required (TallyPrime resolves every master/voucher relative to an
// active company context); envelope_format defaults to json.
func resolveConfig(cfg connectors.RuntimeConfig) (tallyConfig, error) {
	get := func(k string) string { return strings.TrimSpace(cfg.Config[k]) }

	gatewayURL := get("gateway_url")
	if gatewayURL == "" {
		gatewayURL = defaultGatewayURL
	}
	if err := validateGatewayURL(gatewayURL); err != nil {
		return tallyConfig{}, err
	}

	company := get("company")
	if company == "" {
		return tallyConfig{}, errors.New("tally-prime connector requires config company")
	}

	format := strings.ToLower(get("envelope_format"))
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "xml" {
		return tallyConfig{}, fmt.Errorf("tally-prime config envelope_format %q must be json or xml", format)
	}

	return tallyConfig{
		gatewayURL:     gatewayURL,
		company:        company,
		envelopeFormat: format,
		fromDate:       get("from_date"),
		toDate:         get("to_date"),
		timeout:        httpTimeout(cfg),
	}, nil
}

// validateGatewayURL requires an http(s) URL with a host, bounding the
// connector to well-formed loopback/LAN addresses (TallyPrime's Gateway
// Server is never a public internet endpoint; see docs.md's Write actions &
// risks).
func validateGatewayURL(raw string) error {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return fmt.Errorf("tally-prime config gateway_url %q must start with http:// or https://", raw)
	}
	rest := strings.TrimPrefix(strings.TrimPrefix(raw, "http://"), "https://")
	if rest == "" || strings.HasPrefix(rest, "/") {
		return fmt.Errorf("tally-prime config gateway_url %q is missing a host", raw)
	}
	return nil
}

// httpTimeout parses config http_timeout_seconds (default
// defaultHTTPTimeout; invalid/non-positive values fall back to the default
// rather than erroring, since this only bounds a client-side timeout, not a
// protocol correctness input).
func httpTimeout(cfg connectors.RuntimeConfig) time.Duration {
	raw := strings.TrimSpace(cfg.Config["http_timeout_seconds"])
	if raw == "" {
		return defaultHTTPTimeout
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return defaultHTTPTimeout
	}
	return time.Duration(seconds) * time.Second
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Check verifies connection config and, outside fixture mode, POSTs a
// companies Export/Collection envelope to confirm the Gateway Server is
// reachable and the configured company resolves. Fixture mode validates
// config shape only (no network).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}

	def, ok := collectionDefs["companies"]
	if !ok {
		return errors.New("tally-prime: internal error: companies collection definition missing")
	}
	if _, err := c.postCollection(ctx, cfg, conn, def); err != nil {
		return fmt.Errorf("check tally-prime: %w", err)
	}
	return nil
}

// envelopeRequest is the shared ENVELOPE/HEADER/BODY skeleton for a TDL
// Export/Collection request, marshaled to either JSON or XML depending on
// conn.envelopeFormat. Field order/names mirror TallyPrime's own XML
// vocabulary (docs.md's Streams notes example) exactly, since the JSON mode
// is a direct structural mirror of the XML envelope, not an independent
// schema (TallyPrime 7.0+'s native JSON export re-uses the same
// ENVELOPE/HEADER/BODY nesting, only the wire encoding changes).
type envelopeRequest struct {
	XMLName xml.Name       `xml:"ENVELOPE" json:"-"`
	Header  envelopeHeader `xml:"HEADER" json:"HEADER"`
	Body    envelopeBody   `xml:"BODY" json:"BODY"`
}

type envelopeHeader struct {
	Version      string `xml:"VERSION" json:"VERSION"`
	TallyRequest string `xml:"TALLYREQUEST" json:"TALLYREQUEST"`
	Type         string `xml:"TYPE" json:"TYPE"`
	ID           string `xml:"ID" json:"ID"`
}

type envelopeBody struct {
	Desc envelopeDesc `xml:"DESC" json:"DESC"`
}

type envelopeDesc struct {
	StaticVariables staticVariables `xml:"STATICVARIABLES" json:"STATICVARIABLES"`
	TDL             tdlMessage      `xml:"TDL" json:"TDL"`
}

type staticVariables struct {
	SVExportFormat   string `xml:"SVEXPORTFORMAT" json:"SVEXPORTFORMAT"`
	SVCurrentCompany string `xml:"SVCURRENTCOMPANY" json:"SVCURRENTCOMPANY"`
	SVFromDate       string `xml:"SVFROMDATE,omitempty" json:"SVFROMDATE,omitempty"`
	SVToDate         string `xml:"SVTODATE,omitempty" json:"SVTODATE,omitempty"`
}

type tdlMessage struct {
	Message tdlCollectionMessage `xml:"TDLMESSAGE" json:"TDLMESSAGE"`
}

type tdlCollectionMessage struct {
	Collection tdlCollection `xml:"COLLECTION" json:"COLLECTION"`
}

type tdlCollection struct {
	Name     string `xml:"NAME,attr" json:"_NAME"`
	IsModify string `xml:"ISMODIFY,attr" json:"_ISMODIFY"`
	Type     string `xml:"TYPE" json:"TYPE"`
	Fetch    string `xml:"FETCH" json:"FETCH"`
}

// buildEnvelope constructs the Export/Collection request envelope for one
// collectionDef, scoped to conn.company and (when set) conn.fromDate/toDate.
func buildEnvelope(conn tallyConfig, def collectionDef) envelopeRequest {
	sysName := jsonSysName
	if conn.envelopeFormat == "xml" {
		sysName = xmlSysName
	}
	return envelopeRequest{
		Header: envelopeHeader{
			Version:      "1",
			TallyRequest: "Export",
			Type:         "Collection",
			ID:           def.id,
		},
		Body: envelopeBody{
			Desc: envelopeDesc{
				StaticVariables: staticVariables{
					SVExportFormat:   sysName,
					SVCurrentCompany: conn.company,
					SVFromDate:       conn.fromDate,
					SVToDate:         conn.toDate,
				},
				TDL: tdlMessage{
					Message: tdlCollectionMessage{
						Collection: tdlCollection{
							Name:     def.id,
							IsModify: "No",
							Type:     def.tallyType,
							Fetch:    strings.Join(def.fetch, ", "),
						},
					},
				},
			},
		},
	}
}

// postCollection builds, sends, and reads the raw response body for one
// Export/Collection envelope. It never returns before validating the HTTP
// status.
func (c Connector) postCollection(ctx context.Context, cfg connectors.RuntimeConfig, conn tallyConfig, def collectionDef) ([]byte, error) {
	env := buildEnvelope(conn, def)

	var payload []byte
	var contentType string
	var err error
	if conn.envelopeFormat == "xml" {
		payload, err = xml.Marshal(env)
		contentType = "text/xml"
	} else {
		payload, err = json.Marshal(env)
		contentType = "application/json"
	}
	if err != nil {
		return nil, fmt.Errorf("build tally-prime envelope: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, conn.gatewayURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build tally-prime request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)

	client := c.httpClient(cfg)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send tally-prime request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tally-prime gateway returned %s", resp.Status)
	}
	return body, nil
}
