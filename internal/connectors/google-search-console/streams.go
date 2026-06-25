package googlesearchconsole

import (
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// stringify renders a value (which may be a nested array/object from the sitemap
// warnings/errors fields) as a string for storage in a flat record.
func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

// streamKind distinguishes the two API shapes this connector reads:
//   - metaStream: a GET list endpoint (sites, sitemaps) returning a named array.
//   - analyticsStream: a POST searchAnalytics/query with a fixed dimension set.
type streamKind int

const (
	metaStream streamKind = iota
	analyticsStream
)

// streamDef is the per-stream routing entry. The read path is fully data-driven
// from this table: adding a stream means adding one entry plus a Stream catalog
// definition in gscStreams.
type streamDef struct {
	kind streamKind
	// resource is the GET path segment for metaStream entries, relative to the
	// per-site path (sites -> "sites"; sitemaps -> "sites/{site}/sitemaps").
	resource string
	// recordsPath is the JSON path to the array in a metaStream response.
	recordsPath string
	// perSite is true when a metaStream endpoint is scoped to a single site URL
	// (sitemaps); sites itself is account-scoped.
	perSite bool
	// dimensions are the searchAnalytics dimensions for analyticsStream entries.
	dimensions []string
	// mapMeta flattens a raw metaStream object into a Record.
	mapMeta func(map[string]any) connectors.Record
}

// gscStreamDefs is the per-stream routing table.
var gscStreamDefs = map[string]streamDef{
	"sites": {
		kind:        metaStream,
		resource:    "sites",
		recordsPath: "siteEntry",
		perSite:     false,
		mapMeta:     gscSiteRecord,
	},
	"sitemaps": {
		kind:        metaStream,
		resource:    "sitemaps",
		recordsPath: "sitemap",
		perSite:     true,
		mapMeta:     gscSitemapRecord,
	},
	"search_analytics_by_date": {
		kind:       analyticsStream,
		dimensions: []string{"date"},
	},
	"search_analytics_by_country": {
		kind:       analyticsStream,
		dimensions: []string{"country"},
	},
	"search_analytics_by_device": {
		kind:       analyticsStream,
		dimensions: []string{"device"},
	},
	"search_analytics_by_page": {
		kind:       analyticsStream,
		dimensions: []string{"page"},
	},
	"search_analytics_by_query": {
		kind:       analyticsStream,
		dimensions: []string{"query"},
	},
}

// gscStreams returns the published catalog. Search-analytics streams are keyed by
// site_url, search_type, date plus their non-date dimensions (mirroring Airbyte's
// composite key); sites/sitemaps key by their natural identifier.
func gscStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "sites",
			Description: "Google Search Console site properties accessible to the account.",
			PrimaryKey:  []string{"site_url"},
			Fields: []connectors.Field{
				{Name: "site_url", Type: "string"},
				{Name: "permission_level", Type: "string"},
			},
		},
		{
			Name:        "sitemaps",
			Description: "Sitemaps submitted for each site property.",
			PrimaryKey:  []string{"site_url", "path"},
			Fields: []connectors.Field{
				{Name: "site_url", Type: "string"},
				{Name: "path", Type: "string"},
				{Name: "last_submitted", Type: "string"},
				{Name: "last_downloaded", Type: "string"},
				{Name: "is_pending", Type: "boolean"},
				{Name: "is_sitemaps_index", Type: "boolean"},
				{Name: "type", Type: "string"},
				{Name: "warnings", Type: "string"},
				{Name: "errors", Type: "string"},
			},
		},
		analyticsStreamDef("search_analytics_by_date", "Search analytics aggregated by date.", []string{"site_url", "search_type", "date"}),
		analyticsStreamDef("search_analytics_by_country", "Search analytics aggregated by country and date.", []string{"site_url", "search_type", "date", "country"}),
		analyticsStreamDef("search_analytics_by_device", "Search analytics aggregated by device and date.", []string{"site_url", "search_type", "date", "device"}),
		analyticsStreamDef("search_analytics_by_page", "Search analytics aggregated by page and date.", []string{"site_url", "search_type", "date", "page"}),
		analyticsStreamDef("search_analytics_by_query", "Search analytics aggregated by query and date.", []string{"site_url", "search_type", "date", "query"}),
	}
}

func analyticsStreamDef(name, desc string, pk []string) connectors.Stream {
	fields := []connectors.Field{
		{Name: "site_url", Type: "string"},
		{Name: "search_type", Type: "string"},
		{Name: "date", Type: "string"},
	}
	for _, dim := range gscStreamDefs[name].dimensions {
		if dim == "date" {
			continue
		}
		fields = append(fields, connectors.Field{Name: dim, Type: "string"})
	}
	fields = append(fields,
		connectors.Field{Name: "clicks", Type: "number"},
		connectors.Field{Name: "impressions", Type: "number"},
		connectors.Field{Name: "ctr", Type: "number"},
		connectors.Field{Name: "position", Type: "number"},
	)
	return connectors.Stream{
		Name:         name,
		Description:  desc,
		PrimaryKey:   pk,
		CursorFields: []string{"date"},
		Fields:       fields,
	}
}

func gscSiteRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"site_url":         item["siteUrl"],
		"permission_level": item["permissionLevel"],
	}
}

func gscSitemapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"path":              item["path"],
		"last_submitted":    item["lastSubmitted"],
		"last_downloaded":   item["lastDownloaded"],
		"is_pending":        item["isPending"],
		"is_sitemaps_index": item["isSitemapsIndex"],
		"type":              item["type"],
		"warnings":          stringify(item["warnings"]),
		"errors":            stringify(item["errors"]),
	}
}
