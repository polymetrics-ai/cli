package linkedinpages

import (
	"net/url"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to the LinkedIn Pages (Community Management)
// API resource it reads from. LinkedIn Pages streams are heterogeneous:
//   - organizations: a single object at /organizations/{org_id}.
//   - follower_statistics / share_statistics: finder endpoints scoped by the
//     organizationalEntity URN, returning {"elements":[...]} paged with
//     start/count offset parameters.
//   - total_follower_count: a single object at /networkSizes/{urn}.
//
// The routing table captures the resource path, the JSON path to records, the
// finder query (q + extra params), whether the path is org-scoped (suffixes the
// org urn/id), whether the entity is carried in the query, whether the endpoint
// is paged, and the record mapper.
type streamEndpoint struct {
	// resource is the LinkedIn REST resource path segment (e.g.
	// "organizationalEntityFollowerStatistics").
	resource string
	// recordsPath is the dotted JSON path to the records array/object
	// (e.g. "elements"; "" selects the root single object).
	recordsPath string
	// finder is the restli finder name supplied as q (e.g.
	// "organizationalEntity"); empty for plain GET-by-id endpoints.
	finder string
	// entityInQuery, when true, adds organizationalEntity=<org urn> to the query.
	entityInQuery bool
	// entityInPath, when true, suffixes the org urn to the resource path
	// (e.g. networkSizes/urn:li:organization:123).
	entityInPath bool
	// idInPath, when true, suffixes the bare org_id to the resource path
	// (e.g. organizations/123).
	idInPath bool
	// extraQuery are static query params merged into every request.
	extraQuery url.Values
	// paged, when true, drives start/count offset pagination over elements[].
	paged bool
	// mapRecord flattens a raw LinkedIn object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// linkedinPagesStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in linkedinPagesStreams.
var linkedinPagesStreamEndpoints = map[string]streamEndpoint{
	"organizations": {
		resource:    "organizations",
		recordsPath: "",
		idInPath:    true,
		mapRecord:   organizationRecord,
	},
	"follower_statistics": {
		resource:      "organizationalEntityFollowerStatistics",
		recordsPath:   "elements",
		finder:        "organizationalEntity",
		entityInQuery: true,
		paged:         true,
		mapRecord:     followerStatisticsRecord,
	},
	"share_statistics": {
		resource:      "organizationalEntityShareStatistics",
		recordsPath:   "elements",
		finder:        "organizationalEntity",
		entityInQuery: true,
		paged:         true,
		mapRecord:     shareStatisticsRecord,
	},
	"total_follower_count": {
		resource:     "networkSizes",
		recordsPath:  "",
		entityInPath: true,
		extraQuery:   url.Values{"edgeType": []string{"COMPANY_FOLLOWED_BY_MEMBER"}},
		mapRecord:    totalFollowerCountRecord,
	},
}

// linkedinPagesStreams returns the connector's published stream catalog. All
// streams are scoped to a single organization (org_id), so the primary key is
// composed with org_id plus the natural key of each shape.
func linkedinPagesStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "The LinkedIn organization (company page) profile for the configured org_id.",
			PrimaryKey:  []string{"id"},
			Fields:      organizationFields(),
		},
		{
			Name:        "follower_statistics",
			Description: "Lifetime follower statistics for the organization, broken down by association type, seniority, industry, function, staff count range, region, and country.",
			PrimaryKey:  []string{"org_id"},
			Fields:      followerStatisticsFields(),
		},
		{
			Name:        "share_statistics",
			Description: "Lifetime share (content) statistics for the organization: impressions, clicks, likes, comments, shares, and engagement.",
			PrimaryKey:  []string{"org_id"},
			Fields:      shareStatisticsFields(),
		},
		{
			Name:        "total_follower_count",
			Description: "Total first-degree follower count for the organization.",
			PrimaryKey:  []string{"org_id"},
			Fields:      totalFollowerCountFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "org_id", Type: "string"},
		{Name: "urn", Type: "string"},
		{Name: "vanity_name", Type: "string"},
		{Name: "localized_name", Type: "string"},
		{Name: "localized_website", Type: "string"},
		{Name: "primary_organization_type", Type: "string"},
		{Name: "organization_type", Type: "string"},
		{Name: "version_tag", Type: "string"},
		{Name: "staff_count_range", Type: "string"},
		{Name: "name", Type: "object"},
		{Name: "locations", Type: "array"},
		{Name: "industries", Type: "array"},
	}
}

func followerStatisticsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "org_id", Type: "string"},
		{Name: "organizationalEntity", Type: "string"},
		{Name: "followerCountsByAssociationType", Type: "array"},
		{Name: "followerCountsBySeniority", Type: "array"},
		{Name: "followerCountsByIndustry", Type: "array"},
		{Name: "followerCountsByFunction", Type: "array"},
		{Name: "followerCountsByStaffCountRange", Type: "array"},
		{Name: "followerCountsByRegion", Type: "array"},
		{Name: "followerCountsByCountry", Type: "array"},
		{Name: "followerGains", Type: "object"},
	}
}

func shareStatisticsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "org_id", Type: "string"},
		{Name: "organizationalEntity", Type: "string"},
		{Name: "totalShareStatistics", Type: "object"},
		{Name: "shareStatisticsByPost", Type: "array"},
	}
}

func totalFollowerCountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "org_id", Type: "string"},
		{Name: "first_degree_size", Type: "integer"},
	}
}

// organizationRecord flattens the Organization Lookup response. The id is numeric
// and the $URN field carries the organization urn.
func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                        item["id"],
		"urn":                       item["$URN"],
		"vanity_name":               item["vanityName"],
		"localized_name":            item["localizedName"],
		"localized_website":         item["localizedWebsite"],
		"primary_organization_type": item["primaryOrganizationType"],
		"organization_type":         item["organizationType"],
		"version_tag":               item["versionTag"],
		"staff_count_range":         item["staffCountRange"],
		"name":                      item["name"],
		"locations":                 item["locations"],
		"industries":                item["industries"],
	}
}

// followerStatisticsRecord passes through the follower-statistics element. The
// breakdown arrays vary by element, so the full element is preserved alongside
// the entity urn.
func followerStatisticsRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"organizationalEntity": item["organizationalEntity"],
	}
	for _, key := range []string{
		"followerCountsByAssociationType",
		"followerCountsBySeniority",
		"followerCountsByIndustry",
		"followerCountsByFunction",
		"followerCountsByStaffCountRange",
		"followerCountsByRegion",
		"followerCountsByCountry",
		"followerGains",
	} {
		if v, ok := item[key]; ok {
			rec[key] = v
		}
	}
	return rec
}

// shareStatisticsRecord passes through the share-statistics element.
func shareStatisticsRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"organizationalEntity": item["organizationalEntity"],
	}
	for _, key := range []string{"totalShareStatistics", "shareStatisticsByPost"} {
		if v, ok := item[key]; ok {
			rec[key] = v
		}
	}
	return rec
}

// totalFollowerCountRecord flattens the networkSizes single object.
func totalFollowerCountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"first_degree_size": item["firstDegreeSize"],
	}
}
