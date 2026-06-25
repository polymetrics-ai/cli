package strava

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Strava API resource path (relative to
// base_url) and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Strava endpoint path segment (e.g. "athlete/activities").
	// The literal "{athlete_id}" placeholder is substituted at read time from the
	// athlete_id config.
	resource string
	// list is true for endpoints that return a top-level JSON array and support
	// page/per_page pagination; false for single-object endpoints.
	list bool
	// mapRecord flattens a raw Strava object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// stravaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in stravaStreams; the read path
// is fully data-driven from this table.
var stravaStreamEndpoints = map[string]streamEndpoint{
	"activities":    {resource: "athlete/activities", list: true, mapRecord: stravaActivityRecord},
	"athlete":       {resource: "athlete", list: false, mapRecord: stravaAthleteRecord},
	"athlete_stats": {resource: "athletes/{athlete_id}/stats", list: false, mapRecord: stravaStatsRecord},
	"clubs":         {resource: "athlete/clubs", list: true, mapRecord: stravaClubRecord},
}

// stravaStreams returns the connector's published stream catalog.
func stravaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "activities",
			Description:  "Activities owned by the authenticated athlete.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"start_date"},
			Fields:       stravaActivityFields(),
		},
		{
			Name:        "athlete",
			Description: "The authenticated athlete's profile.",
			PrimaryKey:  []string{"id"},
			Fields:      stravaAthleteFields(),
		},
		{
			Name:        "athlete_stats",
			Description: "Activity totals and records for the authenticated athlete.",
			PrimaryKey:  []string{"id"},
			Fields:      stravaStatsFields(),
		},
		{
			Name:        "clubs",
			Description: "Clubs the authenticated athlete belongs to.",
			PrimaryKey:  []string{"id"},
			Fields:      stravaClubFields(),
		},
	}
}

func stravaActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "sport_type", Type: "string"},
		{Name: "distance", Type: "number"},
		{Name: "moving_time", Type: "integer"},
		{Name: "elapsed_time", Type: "integer"},
		{Name: "total_elevation_gain", Type: "number"},
		{Name: "start_date", Type: "string"},
		{Name: "start_date_local", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "average_speed", Type: "number"},
		{Name: "max_speed", Type: "number"},
		{Name: "kudos_count", Type: "integer"},
		{Name: "achievement_count", Type: "integer"},
	}
}

func stravaAthleteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "username", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "sex", Type: "string"},
		{Name: "weight", Type: "number"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func stravaStatsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "biggest_ride_distance", Type: "number"},
		{Name: "biggest_climb_elevation_gain", Type: "number"},
		{Name: "recent_ride_totals", Type: "object"},
		{Name: "recent_run_totals", Type: "object"},
		{Name: "recent_swim_totals", Type: "object"},
		{Name: "ytd_ride_totals", Type: "object"},
		{Name: "ytd_run_totals", Type: "object"},
		{Name: "ytd_swim_totals", Type: "object"},
		{Name: "all_ride_totals", Type: "object"},
		{Name: "all_run_totals", Type: "object"},
		{Name: "all_swim_totals", Type: "object"},
	}
}

func stravaClubFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "sport_type", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "member_count", Type: "integer"},
		{Name: "private", Type: "boolean"},
		{Name: "membership", Type: "string"},
		{Name: "url", Type: "string"},
	}
}

func stravaActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"name":                 item["name"],
		"type":                 item["type"],
		"sport_type":           item["sport_type"],
		"distance":             item["distance"],
		"moving_time":          item["moving_time"],
		"elapsed_time":         item["elapsed_time"],
		"total_elevation_gain": item["total_elevation_gain"],
		"start_date":           item["start_date"],
		"start_date_local":     item["start_date_local"],
		"timezone":             item["timezone"],
		"average_speed":        item["average_speed"],
		"max_speed":            item["max_speed"],
		"kudos_count":          item["kudos_count"],
		"achievement_count":    item["achievement_count"],
	}
}

func stravaAthleteRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"username":   item["username"],
		"firstname":  item["firstname"],
		"lastname":   item["lastname"],
		"city":       item["city"],
		"state":      item["state"],
		"country":    item["country"],
		"sex":        item["sex"],
		"weight":     item["weight"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func stravaStatsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                           item["id"],
		"biggest_ride_distance":        item["biggest_ride_distance"],
		"biggest_climb_elevation_gain": item["biggest_climb_elevation_gain"],
		"recent_ride_totals":           item["recent_ride_totals"],
		"recent_run_totals":            item["recent_run_totals"],
		"recent_swim_totals":           item["recent_swim_totals"],
		"ytd_ride_totals":              item["ytd_ride_totals"],
		"ytd_run_totals":               item["ytd_run_totals"],
		"ytd_swim_totals":              item["ytd_swim_totals"],
		"all_ride_totals":              item["all_ride_totals"],
		"all_run_totals":               item["all_run_totals"],
		"all_swim_totals":              item["all_swim_totals"],
	}
}

func stravaClubRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"sport_type":   item["sport_type"],
		"city":         item["city"],
		"state":        item["state"],
		"country":      item["country"],
		"member_count": item["member_count"],
		"private":      item["private"],
		"membership":   item["membership"],
		"url":          item["url"],
	}
}
