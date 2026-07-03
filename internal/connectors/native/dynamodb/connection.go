package dynamodb

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

const (
	awsService          = "dynamodb"
	scanTarget          = "DynamoDB_20120810.Scan"
	amzJSONContentType  = "application/x-amz-json-1.0"
	defaultReadPageSize = 100
	defaultMaxPages     = 100
	requesterUserAgent  = "polymetrics-go-cli"
)

// connConfig is the validated DynamoDB connection configuration. Secrets
// live in dedicated fields and are never logged; the endpoint is validated
// the same way native/postgres validates its host, bounding SSRF risk from
// operator-supplied config.
type connConfig struct {
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

// resolveConfig validates config + secrets into a connConfig, ported
// rule-for-rule from legacy internal/connectors/dynamodb/dynamodb.go's
// endpoint/requireCredentials helpers. It never logs the secret access key.
func resolveConfig(cfg connectors.RuntimeConfig) (connConfig, error) {
	endpoint, err := resolveEndpoint(cfg)
	if err != nil {
		return connConfig{}, err
	}

	region := strings.TrimSpace(cfg.Config["region"])
	if region == "" {
		return connConfig{}, errors.New("dynamodb connector requires config region")
	}

	accessKeyID := secretValue(cfg, "access_key_id")
	if strings.TrimSpace(accessKeyID) == "" {
		return connConfig{}, errors.New("dynamodb connector requires secret access_key_id")
	}
	secretAccessKey := secretValue(cfg, "secret_access_key")
	if strings.TrimSpace(secretAccessKey) == "" {
		return connConfig{}, errors.New("dynamodb connector requires secret secret_access_key")
	}

	return connConfig{
		endpoint:        endpoint,
		region:          region,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
	}, nil
}

// resolveEndpoint reads config endpoint (falling back to base_url, matching
// legacy's identical fallback) and validates it is an absolute http/https
// URL with a host, bounding SSRF risk from a malformed/attacker-supplied
// endpoint override.
func resolveEndpoint(cfg connectors.RuntimeConfig) (string, error) {
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

// secretValue reads a secret key, tolerating a nil Secrets map.
func secretValue(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// tableName resolves the target table from config table_name (falling back
// to table, matching legacy's identical fallback).
func tableName(cfg connectors.RuntimeConfig) string {
	if table := strings.TrimSpace(cfg.Config["table_name"]); table != "" {
		return table
	}
	return strings.TrimSpace(cfg.Config["table"])
}

// intConfig parses a positive-integer config value, defaulting to fallback
// when unset. Ported verbatim from legacy's intConfig.
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

// fixtureMode reports whether cfg requests the credential-free fixture
// runtime mode (mirrors every other bundle's identical convention).
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Check verifies the connector is configured well enough to talk to
// DynamoDB. In fixture mode it short-circuits without a network call,
// matching legacy's Check exactly (legacy never issued a live Scan from
// Check either — only config/credential validation).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := resolveConfig(cfg); err != nil {
		return err
	}
	return nil
}

// sign attaches AWS Signature Version 4 authentication headers to req,
// ported rule-for-rule from legacy dynamodb.go's sign method: a
// canonicalized request (method, URI, query, signed headers, payload hash)
// is hashed and HMAC-chained through a date/region/service-scoped signing
// key, and the resulting signature is set on the Authorization header. The
// secret access key is used only to derive the signing key HMAC chain; it
// is never placed in a header or logged.
func (c Connector) sign(req *http.Request, conn connConfig, payload []byte, now time.Time) {
	now = now.UTC()
	date := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Host", req.URL.Host)

	payloadHash := sha256Hex(payload)
	canonicalHeaders, signedHeaders := canonicalHeaders(req.Header, req.URL.Host)
	canonicalRequest := strings.Join([]string{
		req.Method, canonicalURI(req.URL), canonicalQuery(req.URL), canonicalHeaders, signedHeaders, payloadHash,
	}, "\n")
	scope := date + "/" + conn.region + "/" + awsService + "/aws4_request"
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256", amzDate, scope, sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := awsSigningKey(conn.secretAccessKey, date, conn.region, awsService)
	sig := hmacSHA256Hex(signingKey, stringToSign)
	req.Header.Set("Authorization",
		"AWS4-HMAC-SHA256 Credential="+conn.accessKeyID+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+sig)
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
