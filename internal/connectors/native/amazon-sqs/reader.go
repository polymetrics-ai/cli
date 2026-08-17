package amazonsqs

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"polymetrics.ai/internal/connectors"
)

// Read performs a bounded ReceiveMessage poll loop, matching legacy's Read
// exactly (amazon_sqs.go:73-107): up to max_polls calls, stopping early the
// first time a poll returns zero messages. Fixture mode emits two canned
// messages without any network access.
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

// receiveForm builds the ReceiveMessage form body. Ported rule-for-rule
// from legacy's receiveForm (amazon_sqs.go:113-129).
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

// messageRecord maps one decoded sqsMessage into a connectors.Record.
// Ported rule-for-rule from legacy's messageRecord (amazon_sqs.go:271-280).
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

// readFixture emits two canned messages without any network access,
// matching legacy's readFixture exactly (amazon_sqs.go:290-300).
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
