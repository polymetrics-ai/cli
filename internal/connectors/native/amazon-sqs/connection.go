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
	queueURL := strings.TrimSpace(cfg.Config["queue_url"])
	if queueURL == "" {
		return sqsConfig{}, errors.New("amazon-sqs connector requires config queue_url")
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
	return sqsConfig{
		queueURL:     queueURL,
		region:       region,
		accessKey:    accessKey,
		secretKey:    secretKey,
		sessionToken: strings.TrimSpace(cfg.Secrets["session_token"]),
	}, nil
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
	_, err := c.do(ctx, cfg, form)
	if err != nil {
		return fmt.Errorf("check amazon-sqs: %w", err)
	}
	return nil
}

// do builds, signs (SigV4), sends, and reads the body of one SQS Query API
// POST. Ported rule-for-rule from legacy's do (amazon_sqs.go:131-171).
func (c Connector) do(ctx context.Context, cfg connectors.RuntimeConfig, form url.Values) ([]byte, error) {
	conn, err := resolveConnConfig(cfg)
	if err != nil {
		return nil, err
	}
	body := []byte(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, conn.queueURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build sqs request: %w", err)
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
		return nil, fmt.Errorf("send sqs request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sqs returned %s", resp.Status)
	}
	return respBody, nil
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
