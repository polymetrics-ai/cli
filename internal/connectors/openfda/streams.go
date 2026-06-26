package openfda

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the openFDA endpoint path (relative to
// base_url, e.g. "drug/event.json"), its stable primary-key field, and the
// record mapper that projects the raw result object into a connectors.Record.
type streamEndpoint struct {
	// path is the openFDA endpoint path segment (e.g. "drug/event.json").
	path string
	// primaryKey is the result field that uniquely identifies a record.
	primaryKey string
	// mapRecord projects a raw openFDA result object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// openfdaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in openfdaStreams; the read
// path is fully data-driven from this table.
var openfdaStreamEndpoints = map[string]streamEndpoint{
	"drug_event":       {path: "drug/event.json", primaryKey: "safetyreportid", mapRecord: drugEventRecord},
	"drug_label":       {path: "drug/label.json", primaryKey: "id", mapRecord: drugLabelRecord},
	"drug_enforcement": {path: "drug/enforcement.json", primaryKey: "recall_number", mapRecord: enforcementRecord},
	"device_event":     {path: "device/event.json", primaryKey: "mdr_report_key", mapRecord: deviceEventRecord},
	"food_enforcement": {path: "food/enforcement.json", primaryKey: "recall_number", mapRecord: enforcementRecord},
}

// openfdaStreams returns the connector's published stream catalog. openFDA
// endpoints are full-refresh, offset-paginated (skip/limit) public datasets, so
// no incremental cursor is published; each stream advertises its stable
// primary-key field.
func openfdaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "drug_event",
			Description: "Adverse drug event reports submitted to the FDA (FAERS).",
			PrimaryKey:  []string{"safetyreportid"},
			Fields:      drugEventFields(),
		},
		{
			Name:        "drug_label",
			Description: "Structured Product Labeling (SPL) drug labels.",
			PrimaryKey:  []string{"id"},
			Fields:      drugLabelFields(),
		},
		{
			Name:        "drug_enforcement",
			Description: "Drug recall enforcement reports.",
			PrimaryKey:  []string{"recall_number"},
			Fields:      enforcementFields(),
		},
		{
			Name:        "device_event",
			Description: "Medical device adverse event reports (MAUDE).",
			PrimaryKey:  []string{"mdr_report_key"},
			Fields:      deviceEventFields(),
		},
		{
			Name:        "food_enforcement",
			Description: "Food recall enforcement reports.",
			PrimaryKey:  []string{"recall_number"},
			Fields:      enforcementFields(),
		},
	}
}

func drugEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "safetyreportid", Type: "string"},
		{Name: "safetyreportversion", Type: "string"},
		{Name: "receivedate", Type: "string"},
		{Name: "receiptdate", Type: "string"},
		{Name: "serious", Type: "string"},
		{Name: "seriousnessdeath", Type: "string"},
		{Name: "transmissiondate", Type: "string"},
		{Name: "primarysourcecountry", Type: "string"},
		{Name: "occurcountry", Type: "string"},
		{Name: "fulfillexpeditecriteria", Type: "string"},
	}
}

func drugLabelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "set_id", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "effective_time", Type: "string"},
		{Name: "indications_and_usage", Type: "array"},
		{Name: "purpose", Type: "array"},
		{Name: "warnings", Type: "array"},
		{Name: "openfda", Type: "object"},
	}
}

func enforcementFields() []connectors.Field {
	return []connectors.Field{
		{Name: "recall_number", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "classification", Type: "string"},
		{Name: "product_type", Type: "string"},
		{Name: "recalling_firm", Type: "string"},
		{Name: "reason_for_recall", Type: "string"},
		{Name: "report_date", Type: "string"},
		{Name: "recall_initiation_date", Type: "string"},
		{Name: "voluntary_mandated", Type: "string"},
		{Name: "distribution_pattern", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "country", Type: "string"},
	}
}

func deviceEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "mdr_report_key", Type: "string"},
		{Name: "report_number", Type: "string"},
		{Name: "event_type", Type: "string"},
		{Name: "date_received", Type: "string"},
		{Name: "date_of_event", Type: "string"},
		{Name: "report_source_code", Type: "string"},
		{Name: "manufacturer_name", Type: "string"},
		{Name: "product_problem_flag", Type: "string"},
		{Name: "adverse_event_flag", Type: "string"},
	}
}

func drugEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"safetyreportid":          item["safetyreportid"],
		"safetyreportversion":     item["safetyreportversion"],
		"receivedate":             item["receivedate"],
		"receiptdate":             item["receiptdate"],
		"serious":                 item["serious"],
		"seriousnessdeath":        item["seriousnessdeath"],
		"transmissiondate":        item["transmissiondate"],
		"primarysourcecountry":    item["primarysourcecountry"],
		"occurcountry":            item["occurcountry"],
		"fulfillexpeditecriteria": item["fulfillexpeditecriteria"],
	}
}

func drugLabelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"set_id":                item["set_id"],
		"version":               item["version"],
		"effective_time":        item["effective_time"],
		"indications_and_usage": item["indications_and_usage"],
		"purpose":               item["purpose"],
		"warnings":              item["warnings"],
		"openfda":               item["openfda"],
	}
}

func enforcementRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"recall_number":          item["recall_number"],
		"status":                 item["status"],
		"classification":         item["classification"],
		"product_type":           item["product_type"],
		"recalling_firm":         item["recalling_firm"],
		"reason_for_recall":      item["reason_for_recall"],
		"report_date":            item["report_date"],
		"recall_initiation_date": item["recall_initiation_date"],
		"voluntary_mandated":     item["voluntary_mandated"],
		"distribution_pattern":   item["distribution_pattern"],
		"state":                  item["state"],
		"country":                item["country"],
	}
}

func deviceEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"mdr_report_key":       item["mdr_report_key"],
		"report_number":        item["report_number"],
		"event_type":           item["event_type"],
		"date_received":        item["date_received"],
		"date_of_event":        item["date_of_event"],
		"report_source_code":   item["report_source_code"],
		"manufacturer_name":    item["manufacturer_name"],
		"product_problem_flag": item["product_problem_flag"],
		"adverse_event_flag":   item["adverse_event_flag"],
	}
}
