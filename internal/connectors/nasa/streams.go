package nasa

import "polymetrics.ai/internal/connectors"

// recordsPath identifies where in a NASA JSON response the record array lives,
// and how each raw object is flattened into a connectors.Record. The various
// NASA APIs hang off different paths (top-level object, near_earth_objects[],
// photos[], a root array), so each stream carries its own extraction recipe.
type streamSpec struct {
	// resource is the API path (relative to base_url) the stream reads from.
	resource string
	// arrayPath is the dotted JSON path to the records array. "" / "." means the
	// root value itself is the record (single object) or a top-level array.
	arrayPath string
	// paginate selects the pagination strategy. Only neo_browse is page-based;
	// everything else is a single bounded request.
	paginate paginationKind
	// mapRecord flattens a raw NASA object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

type paginationKind int

const (
	// paginateNone fetches a single response (single object or one array).
	paginateNone paginationKind = iota
	// paginatePage walks page.number / page.total_pages (NeoWs browse).
	paginatePage
)

// nasaStreamSpecs is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in nasaStreams; the read path is fully
// data-driven from this table.
var nasaStreamSpecs = map[string]streamSpec{
	"apod":        {resource: "planetary/apod", arrayPath: ".", paginate: paginateNone, mapRecord: apodRecord},
	"neo_feed":    {resource: "neo/rest/v1/feed", arrayPath: "near_earth_objects", paginate: paginateNone, mapRecord: neoRecord},
	"neo_browse":  {resource: "neo/rest/v1/neo/browse", arrayPath: "near_earth_objects", paginate: paginatePage, mapRecord: neoRecord},
	"epic":        {resource: "EPIC/api/natural", arrayPath: ".", paginate: paginateNone, mapRecord: epicRecord},
	"mars_photos": {resource: "mars-photos/api/v1/rovers/curiosity/photos", arrayPath: "photos", paginate: paginateNone, mapRecord: marsPhotoRecord},
}

// nasaStreams returns the connector's published stream catalog across the core
// NASA Open APIs (api.nasa.gov).
func nasaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "apod",
			Description:  "NASA Astronomy Picture of the Day entries.",
			PrimaryKey:   []string{"date"},
			CursorFields: []string{"date"},
			Fields:       apodFields(),
		},
		{
			Name:        "neo_feed",
			Description: "Near-Earth Objects with close-approach data for a date range (NeoWs feed).",
			PrimaryKey:  []string{"id"},
			Fields:      neoFields(),
		},
		{
			Name:        "neo_browse",
			Description: "Browse the overall Near-Earth Object dataset, page by page (NeoWs browse).",
			PrimaryKey:  []string{"id"},
			Fields:      neoFields(),
		},
		{
			Name:        "epic",
			Description: "EPIC natural-color Earth imagery metadata.",
			PrimaryKey:  []string{"identifier"},
			Fields:      epicFields(),
		},
		{
			Name:        "mars_photos",
			Description: "Mars rover photos (Curiosity) taken on a given sol.",
			PrimaryKey:  []string{"id"},
			Fields:      marsPhotoFields(),
		},
	}
}

func apodFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "explanation", Type: "string"},
		{Name: "media_type", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "hdurl", Type: "string"},
		{Name: "thumbnail_url", Type: "string"},
		{Name: "copyright", Type: "string"},
		{Name: "service_version", Type: "string"},
	}
}

func neoFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "neo_reference_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "nasa_jpl_url", Type: "string"},
		{Name: "absolute_magnitude_h", Type: "number"},
		{Name: "is_potentially_hazardous_asteroid", Type: "boolean"},
		{Name: "is_sentry_object", Type: "boolean"},
	}
}

func epicFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "string"},
		{Name: "caption", Type: "string"},
		{Name: "image", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "date", Type: "string"},
	}
}

func marsPhotoFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "sol", Type: "integer"},
		{Name: "img_src", Type: "string"},
		{Name: "earth_date", Type: "string"},
		{Name: "camera", Type: "string"},
		{Name: "rover", Type: "string"},
	}
}

func apodRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"date":            item["date"],
		"title":           item["title"],
		"explanation":     item["explanation"],
		"media_type":      item["media_type"],
		"url":             item["url"],
		"hdurl":           item["hdurl"],
		"thumbnail_url":   item["thumbnail_url"],
		"copyright":       item["copyright"],
		"service_version": item["service_version"],
	}
}

func neoRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                                item["id"],
		"neo_reference_id":                  item["neo_reference_id"],
		"name":                              item["name"],
		"nasa_jpl_url":                      item["nasa_jpl_url"],
		"absolute_magnitude_h":              item["absolute_magnitude_h"],
		"is_potentially_hazardous_asteroid": item["is_potentially_hazardous_asteroid"],
		"is_sentry_object":                  item["is_sentry_object"],
	}
}

func epicRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier": item["identifier"],
		"caption":    item["caption"],
		"image":      item["image"],
		"version":    item["version"],
		"date":       item["date"],
	}
}

// marsPhotoRecord flattens the nested camera/rover objects in a Mars rover photo
// down to their human-readable names so the record stays flat and JSONL-friendly.
func marsPhotoRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"sol":        item["sol"],
		"img_src":    item["img_src"],
		"earth_date": item["earth_date"],
		"camera":     nestedName(item["camera"], "full_name", "name"),
		"rover":      nestedName(item["rover"], "name"),
	}
}

// nestedName pulls the first present key out of a nested object value, returning
// nil when the value is not an object or none of the keys are present.
func nestedName(value any, keys ...string) any {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	for _, key := range keys {
		if v, ok := obj[key]; ok {
			return v
		}
	}
	return nil
}
