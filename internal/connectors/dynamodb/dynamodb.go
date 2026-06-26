package dynamodb

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

const (
	service         = "dynamodb"
	targetScan      = "DynamoDB_20120810.Scan"
	contentType     = "application/x-amz-json-1.0"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("dynamodb", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
	Now    func() time.Time
}

type scanResponse struct {
	Items            []map[string]attributeValue `json:"Items"`
	LastEvaluatedKey map[string]attributeValue   `json:"LastEvaluatedKey"`
}

type scanRequest struct {
	TableName         string                    `json:"TableName"`
	Limit             int                       `json:"Limit,omitempty"`
	ExclusiveStartKey map[string]attributeValue `json:"ExclusiveStartKey,omitempty"`
}

type attributeValue map[string]any

func (Connector) Name() string { return "dynamodb" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "dynamodb", DisplayName: "DynamoDB", IntegrationType: "api", Description: "Reads DynamoDB table items through the AWS JSON HTTP API without an SDK.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := endpoint(cfg); err != nil {
		return err
	}
	if err := requireCredentials(cfg); err != nil {
		return err
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "items", Description: "DynamoDB table items.", Fields: []connectors.Field{{Name: "pk", Type: "string"}}, PrimaryKey: []string{"pk"}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "items"
	}
	if stream != "items" {
		return fmt.Errorf("dynamodb stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	if err := requireCredentials(req.Config); err != nil {
		return err
	}
	table := tableName(req.Config)
	if table == "" {
		return errors.New("dynamodb connector requires config table_name")
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	var startKey map[string]attributeValue
	for page := 0; page < maxPages; page++ {
		resp, err := c.scan(ctx, req.Config, scanRequest{TableName: table, Limit: pageSize, ExclusiveStartKey: startKey})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			if err := emit(flattenItem(item)); err != nil {
				return err
			}
		}
		if len(resp.LastEvaluatedKey) == 0 {
			return nil
		}
		startKey = resp.LastEvaluatedKey
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) scan(ctx context.Context, cfg connectors.RuntimeConfig, body scanRequest) (scanResponse, error) {
	base, err := endpoint(cfg)
	if err != nil {
		return scanResponse{}, err
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return scanResponse{}, fmt.Errorf("encode dynamodb scan: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base, bytes.NewReader(payload))
	if err != nil {
		return scanResponse{}, fmt.Errorf("build dynamodb scan: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Target", targetScan)
	req.Header.Set("User-Agent", userAgent)
	if err := c.sign(req, cfg, payload); err != nil {
		return scanResponse{}, err
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return scanResponse{}, fmt.Errorf("send dynamodb scan: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return scanResponse{}, fmt.Errorf("dynamodb scan returned http %d", resp.StatusCode)
	}
	var out scanResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return scanResponse{}, fmt.Errorf("decode dynamodb scan: %w", err)
	}
	return out, nil
}

func (c Connector) sign(req *http.Request, cfg connectors.RuntimeConfig, payload []byte) error {
	accessKey := secret(cfg, "access_key_id")
	secretKey := secret(cfg, "secret_access_key")
	region := strings.TrimSpace(cfg.Config["region"])
	if region == "" {
		return errors.New("dynamodb connector requires config region")
	}
	now := time.Now().UTC()
	if c.Now != nil {
		now = c.Now().UTC()
	}
	date := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Host", req.URL.Host)
	payloadHash := sha256Hex(payload)
	canonicalHeaders, signedHeaders := canonicalHeaders(req.Header, req.URL.Host)
	canonicalRequest := strings.Join([]string{req.Method, canonicalURI(req.URL), canonicalQuery(req.URL), canonicalHeaders, signedHeaders, payloadHash}, "\n")
	scope := date + "/" + region + "/" + service + "/aws4_request"
	stringToSign := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, sha256Hex([]byte(canonicalRequest))}, "\n")
	signingKey := awsSigningKey(secretKey, date, region, service)
	sig := hmacSHA256Hex(signingKey, stringToSign)
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+accessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+sig)
	return nil
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"pk": fmt.Sprintf("fixture#%d", i), "name": fmt.Sprintf("Fixture %d", i), "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func flattenItem(item map[string]attributeValue) connectors.Record {
	out := connectors.Record{}
	for name, value := range item {
		out[name] = attribute(value)
	}
	return out
}

func attribute(v attributeValue) any {
	for kind, raw := range v {
		switch kind {
		case "S", "N", "B":
			return fmt.Sprintf("%v", raw)
		case "BOOL":
			b, _ := raw.(bool)
			return b
		case "NULL":
			return nil
		case "M":
			m, ok := raw.(map[string]any)
			if !ok {
				return raw
			}
			out := connectors.Record{}
			for k, nested := range m {
				if av, ok := nested.(map[string]any); ok {
					out[k] = attribute(attributeValue(av))
				}
			}
			return out
		case "L":
			list, ok := raw.([]any)
			if !ok {
				return raw
			}
			out := make([]any, 0, len(list))
			for _, item := range list {
				if av, ok := item.(map[string]any); ok {
					out = append(out, attribute(attributeValue(av)))
				}
			}
			return out
		default:
			return raw
		}
	}
	return nil
}

func endpoint(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["endpoint"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["base_url"])
	}
	if base == "" {
		return "", errors.New("dynamodb connector requires config endpoint for live mode")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dynamodb config endpoint is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("dynamodb config endpoint must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("dynamodb config endpoint must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(secret(cfg, "access_key_id")) == "" {
		return errors.New("dynamodb connector requires secret access_key_id")
	}
	if strings.TrimSpace(secret(cfg, "secret_access_key")) == "" {
		return errors.New("dynamodb connector requires secret secret_access_key")
	}
	if strings.TrimSpace(cfg.Config["region"]) == "" {
		return errors.New("dynamodb connector requires config region")
	}
	return nil
}

func tableName(cfg connectors.RuntimeConfig) string {
	if table := strings.TrimSpace(cfg.Config["table_name"]); table != "" {
		return table
	}
	return strings.TrimSpace(cfg.Config["table"])
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("dynamodb config %s must be a positive integer", key)
	}
	return value, nil
}

func canonicalHeaders(h http.Header, host string) (string, string) {
	values := map[string]string{"host": host}
	for k, vs := range h {
		name := strings.ToLower(k)
		if name == "authorization" {
			continue
		}
		parts := make([]string, 0, len(vs))
		for _, v := range vs {
			parts = append(parts, strings.Join(strings.Fields(v), " "))
		}
		values[name] = strings.Join(parts, ",")
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		b.WriteString(name)
		b.WriteByte(':')
		b.WriteString(values[name])
		b.WriteByte('\n')
	}
	return b.String(), strings.Join(names, ";")
}

func canonicalURI(u *url.URL) string {
	if u.EscapedPath() == "" {
		return "/"
	}
	return u.EscapedPath()
}

func canonicalQuery(u *url.URL) string { return u.Query().Encode() }

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func awsSigningKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(data))
	return h.Sum(nil)
}

func hmacSHA256Hex(key []byte, data string) string { return hex.EncodeToString(hmacSHA256(key, data)) }

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
