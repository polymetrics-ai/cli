package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// ReadCDC reads DynamoDB Streams records through the reviewed GetShardIterator
// -> GetRecords lifecycle. It is bounded by page_size/max_pages and emits only
// DynamoDB stream event names INSERT/MODIFY/REMOVE. It never starts a live call
// in fixture mode.
func (c Connector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(req.Config) {
		return emit(connectors.CDCEvent{Operation: "INSERT", Record: connectors.Record{"pk": "fixture#1", "fixture": true}, State: connectors.Record{"fixture": true}})
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	if endpoint, err := resolveEndpoint(req.Config, true); err == nil {
		conn.endpoint = endpoint
	} else {
		return err
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultReadPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	iterator := strings.TrimSpace(req.State["shard_iterator"])
	if iterator == "" {
		iterator, err = c.initialShardIterator(ctx, conn, req.Config)
		if err != nil {
			return err
		}
	}
	for page := 0; page < maxPages && iterator != ""; page++ {
		var out map[string]any
		body := map[string]any{"ShardIterator": iterator, "Limit": pageSize}
		if err := c.doJSON(ctx, conn, streamsTargetPrefix+"GetRecords", body, &out); err != nil {
			return err
		}
		for _, event := range cdcEventsFromRecords(out["Records"]) {
			if err := ctx.Err(); err != nil {
				return err
			}
			event.State = connectors.Record{"shard_iterator": fmt.Sprint(out["NextShardIterator"])}
			if err := emit(event); err != nil {
				return err
			}
		}
		iterator, _ = out["NextShardIterator"].(string)
	}
	return nil
}

func (c Connector) initialShardIterator(ctx context.Context, conn connConfig, cfg connectors.RuntimeConfig) (string, error) {
	streamArn := strings.TrimSpace(cfg.Config["stream_arn"])
	shardID := strings.TrimSpace(cfg.Config["shard_id"])
	if streamArn == "" || shardID == "" {
		return "", fmt.Errorf("dynamodb CDC requires stream_arn and shard_id or a shard_iterator state")
	}
	iteratorType := strings.TrimSpace(cfg.Config["iterator_type"])
	if iteratorType == "" {
		iteratorType = "TRIM_HORIZON"
	}
	body := map[string]any{"StreamArn": streamArn, "ShardId": shardID, "ShardIteratorType": iteratorType}
	if seq := strings.TrimSpace(cfg.Config["sequence_number"]); seq != "" {
		body["SequenceNumber"] = seq
	}
	var out map[string]any
	if err := c.doJSON(ctx, conn, streamsTargetPrefix+"GetShardIterator", body, &out); err != nil {
		return "", err
	}
	iterator, _ := out["ShardIterator"].(string)
	if iterator == "" {
		return "", fmt.Errorf("dynamodb GetShardIterator returned no ShardIterator")
	}
	return iterator, nil
}

func cdcEventsFromRecords(value any) []connectors.CDCEvent {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	events := make([]connectors.CDCEvent, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := strings.ToUpper(strings.TrimSpace(fmt.Sprint(m["eventName"])))
		if name == "" {
			name = "MODIFY"
		}
		image := connectors.Record{"event": m}
		if dynamodb, ok := m["dynamodb"].(map[string]any); ok {
			if name == "REMOVE" {
				image = flattenImage(dynamodb["OldImage"])
			} else {
				image = flattenImage(dynamodb["NewImage"])
			}
		}
		events = append(events, connectors.CDCEvent{Operation: name, Record: image})
	}
	return events
}

func flattenImage(value any) connectors.Record {
	m, ok := value.(map[string]any)
	if !ok {
		return connectors.Record{"image": value}
	}
	av := make(map[string]attributeValue, len(m))
	for k, raw := range m {
		if inner, ok := raw.(map[string]any); ok {
			av[k] = attributeValue(inner)
		}
	}
	return flattenItem(av)
}
