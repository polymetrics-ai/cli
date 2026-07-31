package awscloudtrail

import (
	"strings"

	"polymetrics.ai/internal/connectors"
)

func streams() []connectors.Stream {
	out := make([]connectors.Stream, 0, len(cloudTrailPublishedStreams))
	for _, name := range cloudTrailPublishedStreams {
		action := cloudTrailStreamActions[name]
		out = append(out, connectors.Stream{
			Name:        name,
			Description: "AWS CloudTrail " + action + " read stream using a fixed signed JSON-RPC request.",
			PrimaryKey:  []string{"pm_record_id"},
			Fields:      genericFields(action),
		})
	}
	return out
}

func genericFields(action string) []connectors.Field {
	fields := []connectors.Field{{Name: "pm_record_id", Type: "string"}, {Name: "operation", Type: "string"}}
	if action == "LookupEvents" {
		fields = append(fields,
			connectors.Field{Name: "EventId", Type: "string"},
			connectors.Field{Name: "EventName", Type: "string"},
			connectors.Field{Name: "EventSource", Type: "string"},
			connectors.Field{Name: "EventTime", Type: "integer"},
			connectors.Field{Name: "Username", Type: "string"},
			connectors.Field{Name: "AccessKeyId", Type: "string"},
			connectors.Field{Name: "CloudTrailEvent", Type: "string"},
		)
		return fields
	}
	if strings.HasPrefix(action, "Get") || strings.HasPrefix(action, "List") || action == "DescribeTrails" {
		fields = append(fields, connectors.Field{Name: "Name", Type: "string"}, connectors.Field{Name: "Arn", Type: "string"})
	}
	return fields
}
