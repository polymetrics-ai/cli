package googleanalyticsdataapi

import "polymetrics.ai/internal/connectors"

// reportSpec describes a GA4 runReport "stream": a fixed set of dimensions and
// metrics that are projected onto flat record fields. The GA4 Data API has no
// fixed REST resources; instead each report is a dimension x metric query, so a
// stream is modeled as a canned report spec (mirroring how Airbyte ships preset
// reports plus user-defined custom_reports).
type reportSpec struct {
	// name is the published stream name.
	name string
	// description is the human summary.
	description string
	// dimensions are the GA4 API dimension names requested (and the record keys
	// they populate, since each dimension value maps 1:1 to a field).
	dimensions []string
	// metrics are the GA4 API metric names requested.
	metrics []string
}

// gaReports is the per-stream routing table. Each entry is one runReport call.
// The primary key for every report is its dimension set (a row is uniquely
// identified by its dimension tuple within a property), and the cursor field is
// "date" when the report is dimensioned by date (incremental-friendly), else the
// report is full-refresh only.
var gaReports = map[string]reportSpec{
	"daily_active_users": {
		name:        "daily_active_users",
		description: "Active users, new users, and sessions broken down by day.",
		dimensions:  []string{"date"},
		metrics:     []string{"activeUsers", "newUsers", "sessions"},
	},
	"website_overview": {
		name:        "website_overview",
		description: "Top-line engagement metrics broken down by day.",
		dimensions:  []string{"date"},
		metrics:     []string{"activeUsers", "newUsers", "sessions", "screenPageViews", "averageSessionDuration", "bounceRate"},
	},
	"traffic_sources": {
		name:        "traffic_sources",
		description: "Sessions and users by acquisition source / medium per day.",
		dimensions:  []string{"date", "sessionSource", "sessionMedium"},
		metrics:     []string{"sessions", "activeUsers", "newUsers", "engagedSessions"},
	},
	"devices": {
		name:        "devices",
		description: "Users and sessions by device category, OS, and browser per day.",
		dimensions:  []string{"date", "deviceCategory", "operatingSystem", "browser"},
		metrics:     []string{"activeUsers", "sessions", "screenPageViews"},
	},
	"pages": {
		name:        "pages",
		description: "Page views and engagement by page path and title per day.",
		dimensions:  []string{"date", "pagePath", "pageTitle"},
		metrics:     []string{"screenPageViews", "activeUsers", "averageSessionDuration"},
	},
}

// gaStreamOrder fixes the catalog ordering deterministically.
var gaStreamOrder = []string{
	"daily_active_users",
	"website_overview",
	"traffic_sources",
	"devices",
	"pages",
}

// gaStreams returns the connector's published stream catalog. Every report row
// is keyed by its dimension tuple plus the property_id, and reports dimensioned
// by date carry "date" as the incremental cursor field.
func gaStreams() []connectors.Stream {
	streams := make([]connectors.Stream, 0, len(gaStreamOrder))
	for _, name := range gaStreamOrder {
		spec := gaReports[name]
		streams = append(streams, connectors.Stream{
			Name:         spec.name,
			Description:  spec.description,
			PrimaryKey:   primaryKeyFor(spec),
			CursorFields: cursorFieldsFor(spec),
			Fields:       fieldsFor(spec),
		})
	}
	return streams
}

// primaryKeyFor builds the composite primary key: property_id + every dimension.
// Within a property, a row is uniquely identified by its dimension tuple.
func primaryKeyFor(spec reportSpec) []string {
	pk := make([]string, 0, len(spec.dimensions)+1)
	pk = append(pk, "property_id")
	pk = append(pk, spec.dimensions...)
	return pk
}

// cursorFieldsFor returns ["date"] when the report is dimensioned by date so the
// stream can sync incrementally; otherwise it is full-refresh (no cursor).
func cursorFieldsFor(spec reportSpec) []string {
	for _, d := range spec.dimensions {
		if d == "date" {
			return []string{"date"}
		}
	}
	return nil
}

// fieldsFor builds the field list: property_id, each dimension (string), each
// metric (number).
func fieldsFor(spec reportSpec) []connectors.Field {
	fields := make([]connectors.Field, 0, len(spec.dimensions)+len(spec.metrics)+1)
	fields = append(fields, connectors.Field{Name: "property_id", Type: "string"})
	for _, d := range spec.dimensions {
		fields = append(fields, connectors.Field{Name: d, Type: "string"})
	}
	for _, m := range spec.metrics {
		fields = append(fields, connectors.Field{Name: m, Type: "number"})
	}
	return fields
}
