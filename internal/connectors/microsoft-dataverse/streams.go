package microsoftdataverse

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"accounts":      {resource: "accounts", mapRecord: accountRecord},
	"contacts":      {resource: "contacts", mapRecord: contactRecord},
	"leads":         {resource: "leads", mapRecord: leadRecord},
	"opportunities": {resource: "opportunities", mapRecord: opportunityRecord},
	"systemusers":   {resource: "systemusers", mapRecord: systemUserRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "accounts", Description: "Dataverse accounts.", PrimaryKey: []string{"id"}, Fields: commonFields()},
		{Name: "contacts", Description: "Dataverse contacts.", PrimaryKey: []string{"id"}, Fields: commonFields()},
		{Name: "leads", Description: "Dataverse leads.", PrimaryKey: []string{"id"}, Fields: commonFields()},
		{Name: "opportunities", Description: "Dataverse opportunities.", PrimaryKey: []string{"id"}, Fields: commonFields()},
		{Name: "systemusers", Description: "Dataverse system users.", PrimaryKey: []string{"id"}, Fields: commonFields()},
	}
}

func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "created_on", Type: "timestamp"}, {Name: "modified_on", Type: "timestamp"}}
}
func accountRecord(item map[string]any) connectors.Record {
	return baseRecord(first(item, "accountid"), first(item, "name"), item)
}
func contactRecord(item map[string]any) connectors.Record {
	return baseRecord(first(item, "contactid"), first(item, "fullname", "name"), item)
}
func leadRecord(item map[string]any) connectors.Record {
	return baseRecord(first(item, "leadid"), first(item, "fullname", "subject", "name"), item)
}
func opportunityRecord(item map[string]any) connectors.Record {
	return baseRecord(first(item, "opportunityid"), first(item, "name"), item)
}
func systemUserRecord(item map[string]any) connectors.Record {
	return baseRecord(first(item, "systemuserid"), first(item, "fullname", "name"), item)
}
func baseRecord(id, name any, item map[string]any) connectors.Record {
	return connectors.Record{"id": id, "name": name, "email": first(item, "emailaddress1", "internalemailaddress"), "created_on": first(item, "createdon"), "modified_on": first(item, "modifiedon")}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if item[key] != nil {
			return item[key]
		}
	}
	return nil
}
