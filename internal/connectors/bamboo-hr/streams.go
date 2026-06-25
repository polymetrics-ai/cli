package bamboohr

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the BambooHR API resource path (relative
// to base_url), the dotted JSON path where its record array lives, and the record
// mapper that flattens its objects.
//
// BambooHR responses come in three shapes that connsdk.RecordsAt handles:
//   - employees/directory -> {fields:[...], employees:[...]}  (recordsPath "employees")
//   - meta/time_off/types -> {timeOffTypes:[...]}             (recordsPath "timeOffTypes")
//   - meta/fields, meta/lists -> top-level [ ... ]            (recordsPath "")
type streamEndpoint struct {
	// resource is the API path segment, e.g. "employees/directory".
	resource string
	// recordsPath is the dotted JSON path to the record array ("" = top-level array).
	recordsPath string
	// paginated is true for endpoints the connector pages with page/limit; the
	// flat BambooHR meta endpoints return their full list in one response.
	paginated bool
	// mapRecord flattens a raw BambooHR object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// bambooStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bambooStreams; the read path
// is fully data-driven from this table.
var bambooStreamEndpoints = map[string]streamEndpoint{
	"employees":      {resource: "employees/directory", recordsPath: "employees", paginated: true, mapRecord: bambooEmployeeRecord},
	"meta_fields":    {resource: "meta/fields", recordsPath: "", paginated: false, mapRecord: bambooMetaFieldRecord},
	"meta_lists":     {resource: "meta/lists", recordsPath: "", paginated: false, mapRecord: bambooMetaListRecord},
	"time_off_types": {resource: "meta/time_off/types", recordsPath: "timeOffTypes", paginated: false, mapRecord: bambooTimeOffTypeRecord},
}

// bambooStreams returns the connector's published stream catalog.
func bambooStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "employees",
			Description:  "BambooHR employee directory records.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       bambooEmployeeFields(),
		},
		{
			Name:         "meta_fields",
			Description:  "BambooHR account field definitions (meta/fields).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       bambooMetaFieldFields(),
		},
		{
			Name:         "meta_lists",
			Description:  "BambooHR list field definitions and their options (meta/lists).",
			PrimaryKey:   []string{"field_id"},
			CursorFields: nil,
			Fields:       bambooMetaListFields(),
		},
		{
			Name:         "time_off_types",
			Description:  "BambooHR time off type definitions (meta/time_off/types).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       bambooTimeOffTypeFields(),
		},
	}
}

func bambooEmployeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "preferred_name", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "department", Type: "string"},
		{Name: "division", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "work_email", Type: "string"},
		{Name: "work_phone", Type: "string"},
		{Name: "mobile_phone", Type: "string"},
		{Name: "supervisor", Type: "string"},
		{Name: "photo_url", Type: "string"},
	}
}

func bambooMetaFieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "alias", Type: "string"},
		{Name: "deprecated", Type: "boolean"},
	}
}

func bambooMetaListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "field_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "alias", Type: "string"},
		{Name: "manageable", Type: "boolean"},
		{Name: "multiple", Type: "boolean"},
		{Name: "options", Type: "array"},
	}
}

func bambooTimeOffTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "units", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "icon", Type: "string"},
	}
}

func bambooEmployeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             stringField(item, "id"),
		"display_name":   item["displayName"],
		"first_name":     item["firstName"],
		"last_name":      item["lastName"],
		"preferred_name": item["preferredName"],
		"job_title":      item["jobTitle"],
		"department":     item["department"],
		"division":       item["division"],
		"location":       item["location"],
		"work_email":     item["workEmail"],
		"work_phone":     item["workPhone"],
		"mobile_phone":   item["mobilePhone"],
		"supervisor":     item["supervisor"],
		"photo_url":      item["photoUrl"],
	}
}

func bambooMetaFieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         stringField(item, "id"),
		"name":       item["name"],
		"type":       item["type"],
		"alias":      item["alias"],
		"deprecated": item["deprecated"],
	}
}

func bambooMetaListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"field_id":   stringField(item, "fieldId"),
		"name":       item["name"],
		"alias":      item["alias"],
		"manageable": item["manageable"],
		"multiple":   item["multiple"],
		"options":    item["options"],
	}
}

func bambooTimeOffTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    stringField(item, "id"),
		"name":  item["name"],
		"units": item["units"],
		"color": item["color"],
		"icon":  item["icon"],
	}
}
