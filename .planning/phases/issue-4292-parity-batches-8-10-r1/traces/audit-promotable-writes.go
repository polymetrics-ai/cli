// audit-promotable-writes reports the real runtime promotion result for every
// typed write action assigned to #4292. It deliberately calls the engine's
// gate instead of reproducing its record-schema rules in a map generator.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

var assigned = []string{
	"brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio",
	"coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial",
	"datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite",
}

type actionResult struct {
	Action string `json:"action"`
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
}

type connectorResult struct {
	Connector string         `json:"connector"`
	Actions   []actionResult `json:"actions"`
}

type report struct {
	SchemaVersion int               `json:"schema_version"`
	Gate          string            `json:"gate"`
	Rows          []connectorResult `json:"rows"`
}

func main() {
	rows := make([]connectorResult, 0, len(assigned))
	for _, name := range assigned {
		bundle, err := engine.Load(defs.FS, name)
		if err != nil {
			panic(fmt.Sprintf("load %s: %v", name, err))
		}
		row := connectorResult{Connector: name, Actions: make([]actionResult, 0, len(bundle.Writes))}
		for _, action := range bundle.Writes {
			result := actionResult{Action: action.Name, State: "promotable"}
			if err := engine.ValidatePromotableRecordSchema(action.RecordSchema); err != nil {
				result.State = "foundation-gap"
				result.Reason = err.Error()
			}
			row.Actions = append(row.Actions, result)
		}
		sort.Slice(row.Actions, func(i, j int) bool { return row.Actions[i].Action < row.Actions[j].Action })
		rows = append(rows, row)
	}

	raw, err := json.MarshalIndent(report{
		SchemaVersion: 1,
		Gate:          "engine.ValidatePromotableRecordSchema",
		Rows:          rows,
	}, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("encode report: %v", err))
	}
	if err := os.WriteFile(".planning/phases/issue-4292-parity-batches-8-10-r1/PROMOTABLE-WRITE-AUDIT.json", append(raw, '\n'), 0o644); err != nil {
		panic(fmt.Sprintf("write report: %v", err))
	}
}
