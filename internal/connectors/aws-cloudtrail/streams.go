package awscloudtrail

import "polymetrics/internal/connectors"

// streamSpec describes a CloudTrail stream: the LookupEvents read-only filter
// applied to narrow the event set, and the record mapper that flattens an event.
//
// AWS CloudTrail exposes a single read action, LookupEvents, which returns
// management events for the last 90 days. The Airbyte source models distinct
// "streams" as the same LookupEvents call with a different LookupAttributes
// filter (by EventName, EventSource, ReadOnly, etc.). We mirror that: every
// stream maps to the LookupEvents action, differing only by its server-side
// filter so the published catalog reads like familiar per-category event feeds.
type streamSpec struct {
	// filterKey/filterValue, when set, are sent as a LookupAttributes entry on
	// the LookupEvents request body to narrow results server-side.
	filterKey   string
	filterValue string
	mapRecord   func(map[string]any) connectors.Record
}

// streamSpecs is the per-stream routing table. management_events is the
// unfiltered feed; the others are convenience views over the same LookupEvents
// action filtered by AWS attribute.
var streamSpecs = map[string]streamSpec{
	"management_events": {mapRecord: mapEventRecord},
	"read_only_events":  {filterKey: "ReadOnly", filterValue: "true", mapRecord: mapEventRecord},
	"write_only_events": {filterKey: "ReadOnly", filterValue: "false", mapRecord: mapEventRecord},
	"console_logins":    {filterKey: "EventName", filterValue: "ConsoleLogin", mapRecord: mapEventRecord},
}

// streams returns the connector's published stream catalog. Every CloudTrail
// LookupEvents result carries a string EventId and an EventTime timestamp, so
// the primary key is ["EventId"] and the incremental cursor is ["EventTime"].
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "management_events",
			Description:  "AWS CloudTrail management events for the last 90 days via LookupEvents.",
			PrimaryKey:   []string{"EventId"},
			CursorFields: []string{"EventTime"},
			Fields:       eventFields(),
		},
		{
			Name:         "read_only_events",
			Description:  "Management events filtered to ReadOnly=true.",
			PrimaryKey:   []string{"EventId"},
			CursorFields: []string{"EventTime"},
			Fields:       eventFields(),
		},
		{
			Name:         "write_only_events",
			Description:  "Management events filtered to ReadOnly=false (mutating actions).",
			PrimaryKey:   []string{"EventId"},
			CursorFields: []string{"EventTime"},
			Fields:       eventFields(),
		},
		{
			Name:         "console_logins",
			Description:  "Management events filtered to the ConsoleLogin event name.",
			PrimaryKey:   []string{"EventId"},
			CursorFields: []string{"EventTime"},
			Fields:       eventFields(),
		},
	}
}

func eventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "EventId", Type: "string"},
		{Name: "EventName", Type: "string"},
		{Name: "EventSource", Type: "string"},
		{Name: "EventTime", Type: "integer"},
		{Name: "Username", Type: "string"},
		{Name: "AccessKeyId", Type: "string"},
		{Name: "ReadOnly", Type: "string"},
		{Name: "Resources", Type: "array"},
		{Name: "CloudTrailEvent", Type: "string"},
	}
}

// mapEventRecord flattens a LookupEvents result entry into a connectors.Record.
func mapEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"EventId":         item["EventId"],
		"EventName":       item["EventName"],
		"EventSource":     item["EventSource"],
		"EventTime":       item["EventTime"],
		"Username":        item["Username"],
		"AccessKeyId":     item["AccessKeyId"],
		"ReadOnly":        item["ReadOnly"],
		"Resources":       item["Resources"],
		"CloudTrailEvent": item["CloudTrailEvent"],
	}
}
