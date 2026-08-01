package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

func (c Connector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(req.Config) {
		return emit(connectors.CDCEvent{Operation: "INSERT", Record: connectors.Record{"pk": "fixture#1", "fixture": true}, State: connectors.Record{"sequence_number": "fixture-seq-1", "iterator_type": "AFTER_SEQUENCE_NUMBER"}})
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
	if strings.TrimSpace(req.State["sequence_number"]) != "" {
		iterator = ""
	}
	if iterator == "" {
		iterator, err = c.initialShardIterator(ctx, conn, req.Config, req.State)
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
			if event.State == nil {
				event.State = connectors.Record{}
			}
			if streamArn := firstStateOrConfig(req.State, req.Config, "stream_arn"); streamArn != "" {
				event.State["stream_arn"] = streamArn
			}
			if shardID := firstStateOrConfig(req.State, req.Config, "shard_id"); shardID != "" {
				event.State["shard_id"] = shardID
			}
			event.State["iterator_type"] = "AFTER_SEQUENCE_NUMBER"
			if err := emit(event); err != nil {
				return err
			}
		}
		iterator, _ = out["NextShardIterator"].(string)
	}
	return nil
}

func (c Connector) initialShardIterator(ctx context.Context, conn connConfig, cfg connectors.RuntimeConfig, state map[string]string) (string, error) {
	streamArn := firstStateOrConfig(state, cfg, "stream_arn")
	shardID := firstStateOrConfig(state, cfg, "shard_id")
	if streamArn == "" || shardID == "" {
		return "", fmt.Errorf("dynamodb CDC requires stream_arn and shard_id or a shard_iterator state")
	}
	iteratorType := firstStateOrConfig(state, cfg, "iterator_type")
	sequenceNumber := strings.TrimSpace(state["sequence_number"])
	if sequenceNumber != "" {
		iteratorType = "AFTER_SEQUENCE_NUMBER"
	} else {
		sequenceNumber = strings.TrimSpace(cfg.Config["sequence_number"])
	}
	if iteratorType == "" {
		iteratorType = "TRIM_HORIZON"
	}
	body := map[string]any{"StreamArn": streamArn, "ShardId": shardID, "ShardIteratorType": iteratorType}
	if sequenceNumber != "" {
		body["SequenceNumber"] = sequenceNumber
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

func firstStateOrConfig(state map[string]string, cfg connectors.RuntimeConfig, key string) string {
	if value := strings.TrimSpace(state[key]); value != "" {
		return value
	}
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config[key])
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
		sequenceNumber := ""
		if dynamodb, ok := m["dynamodb"].(map[string]any); ok {
			sequenceNumber = strings.TrimSpace(fmt.Sprint(dynamodb["SequenceNumber"]))
			if name == "REMOVE" {
				image = flattenImage(dynamodb["OldImage"])
			} else {
				image = flattenImage(dynamodb["NewImage"])
			}
		}
		state := connectors.Record{}
		if sequenceNumber != "" {
			state["sequence_number"] = sequenceNumber
		}
		events = append(events, connectors.CDCEvent{Operation: name, Record: image, State: state})
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
