package imagga

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Imagga API resource path (relative to
// base_url) it reads from, whether it is a per-image detection endpoint, and the
// mapper that flattens its raw JSON into one or more records.
//
// Imagga's detection endpoints (tags, categories, colors, faces/detections)
// analyze one image per request and return a single result object; the connector
// issues one request per configured image URL and the mapper fans the nested
// result array out into records. The usage endpoint is account-scoped and takes
// no image.
type streamEndpoint struct {
	// resource is the Imagga endpoint path segment (e.g. "tags").
	resource string
	// perImage is true when the endpoint analyzes an image addressed by the
	// image_url query parameter (one request per configured image).
	perImage bool
	// mapRecords flattens a raw Imagga response into zero or more records. imageURL
	// is the image that produced the response ("" for account-scoped endpoints).
	mapRecords func(body map[string]any, imageURL string) []connectors.Record
}

// imaggaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in imaggaStreams.
var imaggaStreamEndpoints = map[string]streamEndpoint{
	"tags":             {resource: "tags", perImage: true, mapRecords: tagsRecords},
	"categories":       {resource: "categories/personal_photos", perImage: true, mapRecords: categoriesRecords},
	"colors":           {resource: "colors", perImage: true, mapRecords: colorsRecords},
	"faces_detections": {resource: "faces/detections", perImage: true, mapRecords: facesRecords},
	"usage":            {resource: "usage", perImage: false, mapRecords: usageRecords},
}

// imaggaStreams returns the connector's published stream catalog. Imagga is a
// full-refresh image-analysis API with no incremental cursor, so streams expose
// no CursorFields.
func imaggaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "tags",
			Description: "Descriptive tags detected per image, one record per tag.",
			PrimaryKey:  []string{"image_url", "tag"},
			Fields:      tagsFields(),
		},
		{
			Name:        "categories",
			Description: "Image categories from the personal_photos categorizer, one record per category.",
			PrimaryKey:  []string{"image_url", "category"},
			Fields:      categoriesFields(),
		},
		{
			Name:        "colors",
			Description: "Dominant colors extracted per image, one record per color.",
			PrimaryKey:  []string{"image_url", "html_code", "color_scope"},
			Fields:      colorsFields(),
		},
		{
			Name:        "faces_detections",
			Description: "Detected face bounding boxes per image, one record per face.",
			PrimaryKey:  []string{"image_url", "face_index"},
			Fields:      facesFields(),
		},
		{
			Name:        "usage",
			Description: "Account API usage totals (monthly, daily, limits).",
			PrimaryKey:  []string{"period"},
			Fields:      usageFields(),
		},
	}
}

func tagsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "image_url", Type: "string"},
		{Name: "tag", Type: "string"},
		{Name: "confidence", Type: "number"},
	}
}

func categoriesFields() []connectors.Field {
	return []connectors.Field{
		{Name: "image_url", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "confidence", Type: "number"},
	}
}

func colorsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "image_url", Type: "string"},
		{Name: "color_scope", Type: "string"},
		{Name: "html_code", Type: "string"},
		{Name: "closest_palette_color", Type: "string"},
		{Name: "percent", Type: "number"},
		{Name: "r", Type: "integer"},
		{Name: "g", Type: "integer"},
		{Name: "b", Type: "integer"},
	}
}

func facesFields() []connectors.Field {
	return []connectors.Field{
		{Name: "image_url", Type: "string"},
		{Name: "face_index", Type: "integer"},
		{Name: "confidence", Type: "number"},
		{Name: "x1", Type: "number"},
		{Name: "y1", Type: "number"},
		{Name: "x2", Type: "number"},
		{Name: "y2", Type: "number"},
	}
}

func usageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "period", Type: "string"},
		{Name: "requests", Type: "integer"},
		{Name: "monthly_processed", Type: "integer"},
		{Name: "monthly_limit", Type: "integer"},
		{Name: "daily_processed", Type: "integer"},
	}
}

// tagsRecords flattens result.tags[] into one record per tag.
func tagsRecords(body map[string]any, imageURL string) []connectors.Record {
	tags := arrayAt(body, "result", "tags")
	out := make([]connectors.Record, 0, len(tags))
	for _, raw := range tags {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, connectors.Record{
			"image_url":  imageURL,
			"tag":        localizedName(item["tag"]),
			"confidence": item["confidence"],
		})
	}
	return out
}

// categoriesRecords flattens result.categories[] into one record per category.
func categoriesRecords(body map[string]any, imageURL string) []connectors.Record {
	cats := arrayAt(body, "result", "categories")
	out := make([]connectors.Record, 0, len(cats))
	for _, raw := range cats {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, connectors.Record{
			"image_url":  imageURL,
			"category":   localizedName(item["name"]),
			"confidence": item["confidence"],
		})
	}
	return out
}

// colorsRecords flattens the overall/foreground/background color groups under
// result.colors into one record per color, tagging each with its scope.
func colorsRecords(body map[string]any, imageURL string) []connectors.Record {
	colors, _ := nestedMap(body, "result", "colors")
	var out []connectors.Record
	for _, scope := range []struct {
		key   string
		label string
	}{
		{"overall_colors", "overall"},
		{"foreground_colors", "foreground"},
		{"background_colors", "background"},
	} {
		group, _ := colors[scope.key].([]any)
		for _, raw := range group {
			item, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			out = append(out, connectors.Record{
				"image_url":             imageURL,
				"color_scope":           scope.label,
				"html_code":             item["html_code"],
				"closest_palette_color": item["closest_palette_color"],
				"percent":               item["percent"],
				"r":                     item["r"],
				"g":                     item["g"],
				"b":                     item["b"],
			})
		}
	}
	return out
}

// facesRecords flattens result.faces[] into one record per detected face.
func facesRecords(body map[string]any, imageURL string) []connectors.Record {
	faces := arrayAt(body, "result", "faces")
	out := make([]connectors.Record, 0, len(faces))
	for i, raw := range faces {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		coords, _ := item["coordinates"].(map[string]any)
		rec := connectors.Record{
			"image_url":  imageURL,
			"face_index": i,
			"confidence": item["confidence"],
		}
		if coords != nil {
			rec["x1"] = coords["xmin"]
			rec["y1"] = coords["ymin"]
			rec["x2"] = coords["xmax"]
			rec["y2"] = coords["ymax"]
		}
		out = append(out, rec)
	}
	return out
}

// usageRecords flattens the account-scoped result object into a single record.
func usageRecords(body map[string]any, _ string) []connectors.Record {
	result, ok := nestedMap(body, "result")
	if !ok {
		return nil
	}
	return []connectors.Record{{
		"period":            "current",
		"requests":          result["total"],
		"monthly_processed": result["monthly_processed"],
		"monthly_limit":     result["monthly_limit"],
		"daily_processed":   result["daily_processed"],
	}}
}

// localizedName extracts the English value from Imagga's {"en": "..."} localized
// name objects, falling back to a plain string.
func localizedName(v any) any {
	switch t := v.(type) {
	case map[string]any:
		if en, ok := t["en"]; ok {
			return en
		}
		return nil
	default:
		return v
	}
}

// nestedMap walks a chain of object keys, returning the map at the end.
func nestedMap(body map[string]any, keys ...string) (map[string]any, bool) {
	cur := body
	for _, k := range keys {
		next, ok := cur[k].(map[string]any)
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

// arrayAt walks object keys then returns the array at the final key.
func arrayAt(body map[string]any, keys ...string) []any {
	if len(keys) == 0 {
		return nil
	}
	parent, ok := nestedMap(body, keys[:len(keys)-1]...)
	if !ok {
		return nil
	}
	arr, _ := parent[keys[len(keys)-1]].([]any)
	return arr
}
