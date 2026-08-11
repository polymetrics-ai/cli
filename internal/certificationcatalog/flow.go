package certificationcatalog

type FlowKind struct {
	ID              string
	SourceRole      string
	DestinationRole string
}

func FlowKinds() []FlowKind {
	return []FlowKind{
		{ID: "api_to_api", SourceRole: "api_source", DestinationRole: "api_destination"},
		{ID: "api_to_database", SourceRole: "api_source", DestinationRole: "database_destination"},
		{ID: "database_to_api", SourceRole: "database_source", DestinationRole: "api_destination"},
		{ID: "database_to_database", SourceRole: "database_source", DestinationRole: "database_destination"},
	}
}
