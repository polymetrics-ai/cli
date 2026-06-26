package googledirectory

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource      string
	recordsPath   string
	customerQuery bool
	mapRecord     func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"users":            {resource: "users", recordsPath: "users", customerQuery: true, mapRecord: userRecord},
	"groups":           {resource: "groups", recordsPath: "groups", customerQuery: true, mapRecord: groupRecord},
	"orgunits":         {resource: "customer/{customer}/orgunits", recordsPath: "organizationUnits", mapRecord: orgUnitRecord},
	"chromeos_devices": {resource: "customer/{customer}/devices/chromeos", recordsPath: "chromeosdevices", mapRecord: chromeDeviceRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "users", Description: "Google Directory users.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "primary_email", Type: "string"}, {Name: "name", Type: "string"}, {Name: "org_unit_path", Type: "string"}}},
		{Name: "groups", Description: "Google Directory groups.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "email", Type: "string"}, {Name: "name", Type: "string"}, {Name: "description", Type: "string"}}},
		{Name: "orgunits", Description: "Google Directory organizational units.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "org_unit_path", Type: "string"}, {Name: "description", Type: "string"}}},
		{Name: "chromeos_devices", Description: "Google Directory ChromeOS devices.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "serial_number", Type: "string"}, {Name: "status", Type: "string"}, {Name: "org_unit_path", Type: "string"}}},
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "primary_email": item["primaryEmail"], "name": nestedString(item, "name", "fullName"), "org_unit_path": item["orgUnitPath"]}
}

func groupRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "email": item["email"], "name": item["name"], "description": item["description"]}
}

func orgUnitRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["orgUnitId"], "name": item["name"], "org_unit_path": item["orgUnitPath"], "description": item["description"]}
}

func chromeDeviceRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["deviceId"], "serial_number": item["serialNumber"], "status": item["status"], "org_unit_path": item["orgUnitPath"]}
}

func nestedString(item map[string]any, objectKey, field string) any {
	obj, ok := item[objectKey].(map[string]any)
	if !ok {
		return nil
	}
	return obj[field]
}
