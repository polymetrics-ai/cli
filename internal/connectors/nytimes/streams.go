package nytimes

import "polymetrics.ai/internal/connectors"

// streamKind classifies how a stream is read from the NYTimes APIs.
type streamKind int

const (
	// kindMostPopular reads a single Most Popular API response keyed by the
	// configured period (1, 7, or 30 days). No pagination.
	kindMostPopular streamKind = iota
	// kindArchive reads the Archive API one month at a time, iterating from
	// start_date to end_date.
	kindArchive
)

// streamDef describes a NYTimes stream: how it is fetched, the API path
// template, the JSON path to the records array, and the record mapper.
type streamDef struct {
	kind streamKind
	// metric is the Most Popular metric segment ("viewed", "emailed",
	// "shared"); empty for the archive stream.
	metric string
	// recordsPath is the dotted JSON path to the array of results.
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

// nytimesStreamDefs is the per-stream routing table; the read path is fully
// data-driven from this map.
var nytimesStreamDefs = map[string]streamDef{
	"most_popular_viewed":  {kind: kindMostPopular, metric: "viewed", recordsPath: "results", mapRecord: mostPopularRecord},
	"most_popular_emailed": {kind: kindMostPopular, metric: "emailed", recordsPath: "results", mapRecord: mostPopularRecord},
	"most_popular_shared":  {kind: kindMostPopular, metric: "shared", recordsPath: "results", mapRecord: mostPopularRecord},
	"archive":              {kind: kindArchive, recordsPath: "response.docs", mapRecord: archiveRecord},
}

// nytimesStreams returns the connector's published stream catalog.
func nytimesStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "most_popular_viewed",
			Description:  "Most viewed NYTimes articles over the configured period.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"published_date"},
			Fields:       mostPopularFields(),
		},
		{
			Name:         "most_popular_emailed",
			Description:  "Most emailed NYTimes articles over the configured period.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"published_date"},
			Fields:       mostPopularFields(),
		},
		{
			Name:         "most_popular_shared",
			Description:  "Most shared NYTimes articles over the configured period (optionally by share_type).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"published_date"},
			Fields:       mostPopularFields(),
		},
		{
			Name:         "archive",
			Description:  "NYTimes article archive, retrieved one month at a time between start_date and end_date.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"pub_date"},
			Fields:       archiveFields(),
		},
	}
}

func mostPopularFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "abstract", Type: "string"},
		{Name: "byline", Type: "string"},
		{Name: "section", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "published_date", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "uri", Type: "string"},
	}
}

func archiveFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "snippet", Type: "string"},
		{Name: "lead_paragraph", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "pub_date", Type: "string"},
		{Name: "document_type", Type: "string"},
		{Name: "news_desk", Type: "string"},
		{Name: "section_name", Type: "string"},
		{Name: "type_of_material", Type: "string"},
		{Name: "word_count", Type: "integer"},
		{Name: "headline", Type: "object"},
	}
}

// mostPopularRecord flattens a Most Popular API result object.
func mostPopularRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"url":            item["url"],
		"title":          item["title"],
		"abstract":       item["abstract"],
		"byline":         item["byline"],
		"section":        item["section"],
		"source":         item["source"],
		"type":           item["type"],
		"published_date": item["published_date"],
		"updated":        item["updated"],
		"uri":            item["uri"],
	}
}

// archiveRecord flattens an Archive API doc object. The API uses "_id" as the
// document identifier; it is surfaced as "id".
func archiveRecord(item map[string]any) connectors.Record {
	id := item["_id"]
	if id == nil {
		id = item["id"]
	}
	return connectors.Record{
		"id":               id,
		"web_url":          item["web_url"],
		"snippet":          item["snippet"],
		"lead_paragraph":   item["lead_paragraph"],
		"source":           item["source"],
		"pub_date":         item["pub_date"],
		"document_type":    item["document_type"],
		"news_desk":        item["news_desk"],
		"section_name":     item["section_name"],
		"type_of_material": item["type_of_material"],
		"word_count":       item["word_count"],
		"headline":         item["headline"],
	}
}
