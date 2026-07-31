package amazonsqs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

const (
	serviceName = "sqs"
	apiVersion  = "2012-11-05"
	userAgent   = "polymetrics-go-cli"
)

// sqsConfig is the validated connection configuration. accessKey/secretKey/
// sessionToken are secrets and are never logged.
type sqsConfig struct {
	queueURL     string
	endpointURL  string
	region       string
	accessKey    string
	secretKey    string
	sessionToken string
}

// resolveConnConfig validates config + secrets into an sqsConfig. Ported
// rule-for-rule from legacy internal/connectors/amazon-sqs/amazon_sqs.go's
// do (amazon_sqs.go:131-144): queue_url and region are required config
// values, access_key/secret_key are required secrets, session_token is an
// optional secret.
func resolveConnConfig(cfg connectors.RuntimeConfig) (sqsConfig, error) {
	queueURL, err := normalizeSQSURL(cfg.Config["queue_url"], "queue_url", true)
	if err != nil {
		return sqsConfig{}, err
	}
	region := strings.TrimSpace(cfg.Config["region"])
	if region == "" {
		return sqsConfig{}, errors.New("amazon-sqs connector requires config region")
	}
	accessKey := strings.TrimSpace(cfg.Secrets["access_key"])
	secretKey := strings.TrimSpace(cfg.Secrets["secret_key"])
	if accessKey == "" || secretKey == "" {
		return sqsConfig{}, errors.New("amazon-sqs connector requires secrets access_key and secret_key")
	}
	endpointURL := strings.TrimSpace(cfg.Config["endpoint_url"])
	if endpointURL == "" {
		endpointURL = serviceEndpointFromQueueURL(queueURL)
	} else {
		endpointURL, err = normalizeSQSURL(endpointURL, "endpoint_url", false)
		if err != nil {
			return sqsConfig{}, err
		}
	}
	return sqsConfig{
		queueURL:     queueURL,
		endpointURL:  endpointURL,
		region:       region,
		accessKey:    accessKey,
		secretKey:    secretKey,
		sessionToken: strings.TrimSpace(cfg.Secrets["session_token"]),
	}, nil
}

func normalizeSQSURL(raw, field string, queueURL bool) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("amazon-sqs connector requires config %s", field)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" || u.Opaque != "" {
		return "", fmt.Errorf("amazon-sqs config %s must be an absolute URL", field)
	}
	if u.User != nil {
		return "", fmt.Errorf("amazon-sqs config %s must not include userinfo", field)
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", fmt.Errorf("amazon-sqs config %s must not include query or fragment components", field)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" {
		if scheme != "http" || !isLocalhost(u.Hostname()) {
			return "", fmt.Errorf("amazon-sqs config %s must use https; http is allowed only for localhost test endpoints", field)
		}
	}
	if !queueURL {
		if u.Path == "" {
			u.Path = "/"
		}
		if u.Path != "/" {
			return "", fmt.Errorf("amazon-sqs config %s must not include a non-root path", field)
		}
	}
	u.Scheme = scheme
	u.RawPath = ""
	return u.String(), nil
}

func isLocalhost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func serviceEndpointFromQueueURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}
	u.Path = "/"
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Check verifies connection config and, outside fixture mode, issues a
// signed GetQueueAttributes call. Fixture mode validates config shape only
// (no network), matching legacy's Check exactly (amazon_sqs.go:51-64).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	form := url.Values{"Action": {"GetQueueAttributes"}, "Version": {apiVersion}, "AttributeName.1": {"QueueArn"}}
	addConfiguredQueueURL(form, cfg)
	_, err := c.do(ctx, cfg, form)
	if err != nil {
		return fmt.Errorf("check amazon-sqs: %w", err)
	}
	return nil
}

// do builds, signs (SigV4), sends, and reads the body of one SQS Query API
// POST. Ported rule-for-rule from legacy's do (amazon_sqs.go:131-171).
func (c Connector) do(ctx context.Context, cfg connectors.RuntimeConfig, form url.Values) ([]byte, error) {
	resp, err := c.doQueue(ctx, cfg, form, 16<<20)
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}

type sqsHTTPResponse struct {
	status int
	body   []byte
}

func (c Connector) doQueue(ctx context.Context, cfg connectors.RuntimeConfig, form url.Values, maxBytes int) (sqsHTTPResponse, error) {
	conn, err := resolveConnConfig(cfg)
	if err != nil {
		return sqsHTTPResponse{}, err
	}
	form = cloneValues(form)
	form.Set("QueueUrl", conn.queueURL)
	return c.doEndpoint(ctx, conn, conn.queueURL, form, maxBytes)
}

func (c Connector) doService(ctx context.Context, cfg connectors.RuntimeConfig, form url.Values, maxBytes int) (sqsHTTPResponse, error) {
	conn, err := resolveConnConfig(cfg)
	if err != nil {
		return sqsHTTPResponse{}, err
	}
	form = cloneValues(form)
	if strings.TrimSpace(form.Get("QueueUrl")) != "" {
		form.Set("QueueUrl", conn.queueURL)
	}
	return c.doEndpoint(ctx, conn, conn.endpointURL, form, maxBytes)
}

func cloneValues(in url.Values) url.Values {
	out := make(url.Values, len(in))
	for k, values := range in {
		out[k] = append([]string(nil), values...)
	}
	return out
}

func (c Connector) doEndpoint(ctx context.Context, conn sqsConfig, endpoint string, form url.Values, maxBytes int) (sqsHTTPResponse, error) {
	if maxBytes <= 0 || maxBytes > 16<<20 {
		maxBytes = 16 << 20
	}
	body := []byte(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return sqsHTTPResponse{}, fmt.Errorf("build sqs request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("User-Agent", userAgent)
	if conn.sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", conn.sessionToken)
	}
	c.sign(req, body, conn.accessKey, conn.secretKey, conn.region)

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return sqsHTTPResponse{}, fmt.Errorf("send sqs request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, int64(maxBytes)+1))
	if readErr != nil {
		return sqsHTTPResponse{}, fmt.Errorf("read sqs response: %w", readErr)
	}
	if len(respBody) > maxBytes {
		return sqsHTTPResponse{}, fmt.Errorf("sqs response too large: %d bytes exceeds limit %d", len(respBody), maxBytes)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return sqsHTTPResponse{}, fmt.Errorf("sqs returned %s", resp.Status)
	}
	return sqsHTTPResponse{status: resp.StatusCode, body: respBody}, nil
}

func addConfiguredQueueURL(form url.Values, cfg connectors.RuntimeConfig) {
	if strings.TrimSpace(form.Get("QueueUrl")) == "" {
		form.Set("QueueUrl", strings.TrimSpace(cfg.Config["queue_url"]))
	}
}

// sign computes and attaches an AWS SigV4 Authorization header. Ported
// rule-for-rule from legacy's sign (amazon_sqs.go:173-189).
func (c Connector) sign(req *http.Request, body []byte, accessKey, secretKey, region string) {
	now := time.Now().UTC()
	if c.Now != nil {
		now = c.Now().UTC()
	}
	amzDate := now.Format("20060102T150405Z")
	date := now.Format("20060102")
	payloadHash := sha256Hex(body)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	signedHeaders, canonicalHeaders := canonicalHeaders(req)
	canonical := strings.Join([]string{req.Method, canonicalURI(req.URL), req.URL.RawQuery, canonicalHeaders, signedHeaders, payloadHash}, "\n")
	scope := date + "/" + region + "/" + serviceName + "/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + sha256Hex([]byte(canonical))
	sig := hex.EncodeToString(hmacSHA256(signingKey(secretKey, date, region), stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+accessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+sig)
}

func canonicalHeaders(req *http.Request) (string, string) {
	values := map[string]string{"host": req.URL.Host}
	for k, vs := range req.Header {
		lk := strings.ToLower(k)
		values[lk] = strings.Join(vs, ",")
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(strings.Join(strings.Fields(values[k]), " "))
		b.WriteByte('\n')
	}
	return strings.Join(keys, ";"), b.String()
}

func canonicalURI(u *url.URL) string {
	if u.EscapedPath() == "" {
		return "/"
	}
	return u.EscapedPath()
}

func signingKey(secret, date, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, serviceName)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
