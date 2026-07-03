package dynamodb

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// attributeValue is one DynamoDB AttributeValue envelope — a single-key map
// naming the wire type (S/N/B/BOOL/NULL/M/L) and carrying its raw value.
// Ported verbatim from legacy dynamodb.go.
type attributeValue map[string]any

// Catalog returns the single static "items" stream, matching legacy's
// hardcoded one-stream catalog (dynamodb.go's Catalog): DynamoDB item shape
// is only knowable per-table at Scan time, so this is a fixed, generic
// catalog entry rather than a discovered-at-runtime shape (capabilities.
// dynamic_schema stays false).
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{
		Connector: c.Name(),
		Streams: []connectors.Stream{
			{
				Name:        itemsStreamName,
				Description: "DynamoDB table items.",
				Fields:      []connectors.Field{{Name: "pk", Type: "string"}},
				PrimaryKey:  []string{"pk"},
			},
		},
	}, nil
}

// flattenItem converts one DynamoDB Scan item (a map of attribute name ->
// AttributeValue envelope) into a plain connectors.Record. Ported verbatim
// from legacy dynamodb.go's flattenItem.
func flattenItem(item map[string]attributeValue) connectors.Record {
	out := connectors.Record{}
	for name, value := range item {
		out[name] = attribute(value)
	}
	return out
}

// attribute recursively unwraps a single DynamoDB AttributeValue envelope
// into a plain Go value: S/N/B all stringify (DynamoDB itself represents
// numbers as decimal strings on the wire), BOOL/NULL map directly, M
// recurses into a nested connectors.Record, and L recurses into a slice of
// unwrapped values (non-object list elements are dropped, mirroring
// connsdk.RecordsAt's own tolerance for non-object array elements
// elsewhere in this engine). An attribute value naming an unrecognized
// type is passed through raw rather than dropped, so an unexpected/future
// DynamoDB type never silently loses data. Ported verbatim from legacy
// dynamodb.go's attribute.
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
			for _, elem := range list {
				if av, ok := elem.(map[string]any); ok {
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
