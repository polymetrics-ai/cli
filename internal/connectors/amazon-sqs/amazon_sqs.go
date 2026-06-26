// Package amazonsqs implements a conservative Amazon SQS source connector.
// It signs Query API requests with AWS SigV4 and only calls ReceiveMessage or
// GetQueueAttributes; it never sends, deletes, or changes messages. ReceiveMessage
// can still affect message visibility according to SQS semantics.
package amazonsqs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"polymetrics.ai/internal/connectors"
)

const (
	connectorName = "amazon-sqs"
	serviceName   = "sqs"
	apiVersion    = "2012-11-05"
	userAgent     = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
	Now    func() time.Time
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Amazon SQS", IntegrationType: "api", Description: "Reads messages from Amazon SQS via signed ReceiveMessage calls. Read-only; messages are not deleted.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

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

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{{Name: "messages", Description: "Messages received from the configured SQS queue. The connector does not delete messages.", PrimaryKey: []string{"message_id"}, Fields: []connectors.Field{{Name: "message_id", Type: "string"}, {Name: "md5_of_body", Type: "string"}, {Name: "body", Type: "object"}, {Name: "sent_timestamp", Type: "string"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "messages"
	}
	if stream != "messages" {
		return fmt.Errorf("amazon-sqs stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	maxPolls := intConfig(req.Config.Config["max_polls"], 1, 1, 100)
	for i := 0; i < maxPolls; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body, err := c.do(ctx, req.Config, receiveForm(req.Config))
		if err != nil {
			return fmt.Errorf("read amazon-sqs messages: %w", err)
		}
		messages, err := parseMessages(body)
		if err != nil {
			return err
		}
		for _, msg := range messages {
			if err := emit(messageRecord(msg)); err != nil {
				return err
			}
		}
		if len(messages) == 0 {
			return nil
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func receiveForm(cfg connectors.RuntimeConfig) url.Values {
	form := url.Values{"Action": {"ReceiveMessage"}, "Version": {apiVersion}}
	form.Set("MaxNumberOfMessages", strconv.Itoa(intConfig(cfg.Config["max_batch_size"], 10, 1, 10)))
	form.Set("WaitTimeSeconds", strconv.Itoa(intConfig(cfg.Config["max_wait_time"], 0, 0, 20)))
	if v := strings.TrimSpace(cfg.Config["visibility_timeout"]); v != "" {
		form.Set("VisibilityTimeout", strconv.Itoa(intConfig(v, 30, 0, 43200)))
	}
	attrs := strings.TrimSpace(cfg.Config["attributes_to_return"])
	if attrs == "" {
		attrs = "All"
	}
	for i, attr := range splitCSV(attrs) {
		form.Set(fmt.Sprintf("MessageAttributeName.%d", i+1), attr)
	}
	form.Set("AttributeName.1", "All")
	return form
}

func (c Connector) do(ctx context.Context, cfg connectors.RuntimeConfig, form url.Values) ([]byte, error) {
	queueURL := strings.TrimSpace(cfg.Config["queue_url"])
	if queueURL == "" {
		return nil, errors.New("amazon-sqs connector requires config queue_url")
	}
	region := strings.TrimSpace(cfg.Config["region"])
	if region == "" {
		return nil, errors.New("amazon-sqs connector requires config region")
	}
	accessKey := strings.TrimSpace(cfg.Secrets["access_key"])
	secretKey := strings.TrimSpace(cfg.Secrets["secret_key"])
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("amazon-sqs connector requires secrets access_key and secret_key")
	}
	body := []byte(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queueURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build sqs request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("User-Agent", userAgent)
	if token := strings.TrimSpace(cfg.Secrets["session_token"]); token != "" {
		req.Header.Set("X-Amz-Security-Token", token)
	}
	c.sign(req, body, accessKey, secretKey, region)
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send sqs request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sqs returned %s", resp.Status)
	}
	return respBody, nil
}

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

type receiveMessageResponse struct {
	Messages []sqsMessage `xml:"ReceiveMessageResult>Message"`
}

type sqsMessage struct {
	MessageID         string        `xml:"MessageId"`
	ReceiptHandle     string        `xml:"ReceiptHandle"`
	MD5OfBody         string        `xml:"MD5OfBody"`
	Body              string        `xml:"Body"`
	Attributes        []nameValue   `xml:"Attribute"`
	MessageAttributes []messageAttr `xml:"MessageAttribute"`
}

type nameValue struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

type messageAttr struct {
	Name  string `xml:"Name"`
	Value struct {
		StringValue string `xml:"StringValue"`
		DataType    string `xml:"DataType"`
	} `xml:"Value"`
}

func parseMessages(body []byte) ([]sqsMessage, error) {
	var out receiveMessageResponse
	if err := xml.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode sqs xml: %w", err)
	}
	return out.Messages, nil
}

func messageRecord(msg sqsMessage) connectors.Record {
	rec := connectors.Record{"message_id": msg.MessageID, "md5_of_body": msg.MD5OfBody, "receipt_handle": msg.ReceiptHandle, "body": parseBody(msg.Body)}
	for _, attr := range msg.Attributes {
		rec[snake(attr.Name)] = attr.Value
	}
	for _, attr := range msg.MessageAttributes {
		rec[snake(attr.Name)] = attr.Value.StringValue
	}
	return rec
}

func parseBody(body string) any {
	var out any
	if err := json.Unmarshal([]byte(body), &out); err == nil {
		return out
	}
	return body
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"message_id": fmt.Sprintf("message_fixture_%d", i), "body": map[string]any{"fixture": true, "n": i}, "sent_timestamp": "1767225600000", "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func splitCSV(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func intConfig(raw string, def, min, max int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		n = def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func snake(s string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if unicode.IsUpper(r) {
			if b.Len() > 0 && !lastUnderscore {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		} else if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
