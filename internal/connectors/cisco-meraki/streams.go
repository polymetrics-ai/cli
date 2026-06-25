package ciscomeraki

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Meraki API resource path it reads from
// and the record mapper that flattens its objects. orgScoped streams require an
// organizationId and are read by fanning out across the organizations list; their
// pathFor closure builds the per-organization path.
type streamEndpoint struct {
	// orgScoped is true when the resource lives under /organizations/{id}/...
	// (networks, devices, admins). false means a top-level resource
	// (/organizations).
	orgScoped bool
	// resource is the top-level path for non-orgScoped streams (e.g.
	// "organizations").
	resource string
	// pathFor builds the per-organization path for orgScoped streams.
	pathFor func(orgID string) string
	// mapRecord flattens a raw Meraki object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// merakiStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in merakiStreams.
var merakiStreamEndpoints = map[string]streamEndpoint{
	"organizations": {
		resource:  "organizations",
		mapRecord: merakiOrganizationRecord,
	},
	"organization_networks": {
		orgScoped: true,
		pathFor:   func(orgID string) string { return "organizations/" + orgID + "/networks" },
		mapRecord: merakiNetworkRecord,
	},
	"organization_devices": {
		orgScoped: true,
		pathFor:   func(orgID string) string { return "organizations/" + orgID + "/devices" },
		mapRecord: merakiDeviceRecord,
	},
	"organization_admins": {
		orgScoped: true,
		pathFor:   func(orgID string) string { return "organizations/" + orgID + "/admins" },
		mapRecord: merakiAdminRecord,
	},
}

// merakiStreams returns the connector's published stream catalog. The Meraki API
// is a configuration/state API with no global incremental cursor field across
// these resources, so the streams are full-refresh with stable primary keys.
func merakiStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "Meraki organizations the API key can access.",
			PrimaryKey:  []string{"id"},
			Fields:      merakiOrganizationFields(),
		},
		{
			Name:        "organization_networks",
			Description: "Networks within each accessible organization.",
			PrimaryKey:  []string{"id"},
			Fields:      merakiNetworkFields(),
		},
		{
			Name:        "organization_devices",
			Description: "Devices managed by each accessible organization.",
			PrimaryKey:  []string{"serial"},
			Fields:      merakiDeviceFields(),
		},
		{
			Name:        "organization_admins",
			Description: "Dashboard administrators of each accessible organization.",
			PrimaryKey:  []string{"id"},
			Fields:      merakiAdminFields(),
		},
	}
}

func merakiOrganizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "api", Type: "object"},
		{Name: "cloud", Type: "object"},
		{Name: "licensing", Type: "object"},
		{Name: "management", Type: "object"},
	}
}

func merakiNetworkFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "organizationId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "productTypes", Type: "array"},
		{Name: "timeZone", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "enrollmentString", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "isBoundToConfigTemplate", Type: "boolean"},
	}
}

func merakiDeviceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "serial", Type: "string"},
		{Name: "organizationId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "mac", Type: "string"},
		{Name: "networkId", Type: "string"},
		{Name: "productType", Type: "string"},
		{Name: "lat", Type: "number"},
		{Name: "lng", Type: "number"},
		{Name: "address", Type: "string"},
		{Name: "firmware", Type: "string"},
		{Name: "tags", Type: "array"},
	}
}

func merakiAdminFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "organizationId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "orgAccess", Type: "string"},
		{Name: "authenticationMethod", Type: "string"},
		{Name: "accountStatus", Type: "string"},
		{Name: "twoFactorAuthEnabled", Type: "boolean"},
		{Name: "hasApiKey", Type: "boolean"},
		{Name: "lastActive", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "networks", Type: "array"},
	}
}

func merakiOrganizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"url":        item["url"],
		"api":        item["api"],
		"cloud":      item["cloud"],
		"licensing":  item["licensing"],
		"management": item["management"],
	}
}

func merakiNetworkRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"organizationId":          item["organizationId"],
		"name":                    item["name"],
		"productTypes":            item["productTypes"],
		"timeZone":                item["timeZone"],
		"tags":                    item["tags"],
		"enrollmentString":        item["enrollmentString"],
		"url":                     item["url"],
		"notes":                   item["notes"],
		"isBoundToConfigTemplate": item["isBoundToConfigTemplate"],
	}
}

func merakiDeviceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"serial":         item["serial"],
		"organizationId": item["organizationId"],
		"name":           item["name"],
		"model":          item["model"],
		"mac":            item["mac"],
		"networkId":      item["networkId"],
		"productType":    item["productType"],
		"lat":            item["lat"],
		"lng":            item["lng"],
		"address":        item["address"],
		"firmware":       item["firmware"],
		"tags":           item["tags"],
	}
}

func merakiAdminRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"organizationId":       item["organizationId"],
		"name":                 item["name"],
		"email":                item["email"],
		"orgAccess":            item["orgAccess"],
		"authenticationMethod": item["authenticationMethod"],
		"accountStatus":        item["accountStatus"],
		"twoFactorAuthEnabled": item["twoFactorAuthEnabled"],
		"hasApiKey":            item["hasApiKey"],
		"lastActive":           item["lastActive"],
		"tags":                 item["tags"],
		"networks":             item["networks"],
	}
}
