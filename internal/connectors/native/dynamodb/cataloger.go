package dynamodb

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// attributeValue is one DynamoDB AttributeValue envelope — a single-key map
// naming the wire type (S/N/B/BOOL/NULL/M/L/SS/NS/BS) and carrying its raw value.
type attributeValue map[string]any

// Catalog returns the reviewed static operation streams from defs/dynamodb.
// DynamoDB item shapes are table-specific, so the item/query streams expose a
// generic pk field while metadata/changefeed streams expose id/operation/
// response fields. No live discovery call is made.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(readOperations))
	for _, op := range readOperations {
		fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "operation", Type: "string"}, {Name: "response", Type: "json"}}
		pk := []string{"id"}
		if op.Stream == itemsStreamName || op.Stream == "query_items" {
			fields = []connectors.Field{{Name: "pk", Type: "string"}, {Name: "operation", Type: "string"}}
			pk = []string{"pk"}
		}
		streams = append(streams, connectors.Stream{Name: op.Stream, Description: op.Description, Fields: fields, PrimaryKey: pk})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

// flattenItem converts one DynamoDB item (attribute name -> AttributeValue
// envelope) into a plain connectors.Record.
func flattenItem(item map[string]attributeValue) connectors.Record {
	out := connectors.Record{}
	for name, value := range item {
		out[name] = attribute(value)
	}
	return out
}

// attribute recursively unwraps a DynamoDB AttributeValue envelope into a
// plain Go value. Unknown/future DynamoDB value kinds are passed through raw
// rather than dropped.
func attribute(v attributeValue) any {
	for kind, raw := range v {
		switch kind {
		case "S", "N", "B":
			return fmt.Sprintf("%v", raw)
		case "SS", "NS", "BS":
			return raw
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
