package dynamodb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// scanRequest is the DynamoDB_20120810.Scan JSON-RPC request body, ported
// verbatim from legacy dynamodb.go.
type scanRequest struct {
	TableName         string                    `json:"TableName"`
	Limit             int                       `json:"Limit,omitempty"`
	ExclusiveStartKey map[string]attributeValue `json:"ExclusiveStartKey,omitempty"`
}

// scanResponse is the DynamoDB_20120810.Scan JSON-RPC response body, ported
// verbatim from legacy dynamodb.go.
type scanResponse struct {
	Items            []map[string]attributeValue `json:"Items"`
	LastEvaluatedKey map[string]attributeValue   `json:"LastEvaluatedKey"`
}

// InitialState satisfies connectors.StatefulReader. DynamoDB Scan has no
// server-side incremental filter (legacy never modeled one either), so a
// stream starts with an empty cursor purely for interface conformity.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read performs a paginated DynamoDB_20120810.Scan over the configured
// table, flattening each returned item into a connectors.Record. Ported
// rule-for-rule from legacy dynamodb.go's Read/Connector.scan.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = itemsStreamName
	}
	if stream != itemsStreamName {
		return fmt.Errorf("dynamodb stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}

	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	table := tableName(req.Config)
	if table == "" {
		return fmt.Errorf("dynamodb connector requires config table_name")
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultReadPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}

	var startKey map[string]attributeValue
	for page := 0; page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := c.scan(ctx, conn, scanRequest{TableName: table, Limit: pageSize, ExclusiveStartKey: startKey})
		if err != nil {
			return err
		}
		for _, item := range resp.Items {
			if err := ctx.Err(); err != nil {
				return err
			}
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

// scan issues one signed DynamoDB_20120810.Scan request and decodes its
// response, ported rule-for-rule from legacy dynamodb.go's Connector.scan.
func (c Connector) scan(ctx context.Context, conn connConfig, body scanRequest) (scanResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return scanResponse{}, fmt.Errorf("encode dynamodb scan: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, conn.endpoint, bytes.NewReader(payload))
	if err != nil {
		return scanResponse{}, fmt.Errorf("build dynamodb scan: %w", err)
	}
	httpReq.Header.Set("Content-Type", amzJSONContentType)
	httpReq.Header.Set("X-Amz-Target", scanTarget)
	httpReq.Header.Set("User-Agent", requesterUserAgent)

	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	c.sign(httpReq, conn, payload, now)

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return scanResponse{}, fmt.Errorf("send dynamodb scan: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
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

// readFixture emits deterministic records without any network access so the
// conformance harness and unit tests can exercise dynamodb credential-free.
// Ported verbatim from legacy dynamodb.go's readFixture.
func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"pk": fmt.Sprintf("fixture#%d", i), "name": fmt.Sprintf("Fixture %d", i), "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
