package hubspot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics/internal/connectors"
)

// writeAction describes an allow-listed reverse-ETL action: the HTTP method and
// the CRM v3 object it targets. update_contact appends the contact id to the
// path at execution time.
type writeAction struct {
	method string
	object string
	// needsID is true when the action mutates an existing object addressed by id.
	needsID bool
}

// hubspotWriteActions is the reverse-ETL allow-list. Anything not present here is
// rejected by ValidateWrite. Starts minimal (contact create/update) to prove the
// write template; expand by adding entries.
var hubspotWriteActions = map[string]writeAction{
	"create_contact": {method: http.MethodPost, object: "contacts"},
	"update_contact": {method: http.MethodPatch, object: "contacts", needsID: true},
}

// contactWritableProperties is the allow-listed set of HubSpot contact
// properties a reverse-ETL write may set. Anything outside this set is dropped so
// the write surface stays bounded.
var contactWritableProperties = []string{"email", "firstname", "lastname", "phone", "company", "lifecyclestage", "website", "jobtitle"}

// ValidateWrite enforces the action allow-list. It never inspects or logs secret
// values; it only confirms the action is approved and records are well-formed.
func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	action, spec, err := resolveWriteAction(req.Action)
	if err != nil {
		return err
	}
	for i, record := range records {
		if err := validateWriteRecord(action, spec, record); err != nil {
			return fmt.Errorf("hubspot %s record %d: %w", action, i+1, err)
		}
	}
	return nil
}

// DryRunWrite validates the request and returns a staged-count preview with no
// network call, supporting the plan->preview->approve->execute reverse-ETL flow.
func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WritePreview{}, err
	}
	action, _, _ := resolveWriteAction(req.Action)
	return connectors.WritePreview{
		RecordsStaged: len(records),
		Action:        action,
		Warnings:      []string{"hubspot write executes a live mutation only after approval; dry run performs no external call"},
	}, nil
}

// Write executes an allow-listed action by sending JSON {properties:{...}}
// payloads to the HubSpot CRM v3 endpoint. Secrets and record PII are never
// logged; errors reference the action and record index only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	action, spec, _ := resolveWriteAction(req.Action)

	if fixtureMode(req.Config) {
		// Fixture mode never mutates the live portal; report a receipt-style
		// success so credential-free conformance can exercise the write path.
		return connectors.WriteResult{RecordsWritten: len(records)}, nil
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}

	result := connectors.WriteResult{}
	for i, record := range records {
		if err := ctx.Err(); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, err
		}
		path := hubspotObjectsBase + "/" + spec.object
		if spec.needsID {
			id := stringField(record, "id")
			if id == "" {
				result.RecordsFailed = len(records) - result.RecordsWritten
				return result, fmt.Errorf("hubspot %s record %d: missing id", action, i+1)
			}
			path = path + "/" + url.PathEscape(id)
		}
		body := map[string]any{"properties": contactProperties(record)}
		if _, err := r.Do(ctx, spec.method, path, nil, body); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("hubspot %s record %d: %w", action, i+1, err)
		}
		result.RecordsWritten++
	}
	return result, nil
}

// resolveWriteAction normalizes and looks up an action against the allow-list.
func resolveWriteAction(raw string) (string, writeAction, error) {
	action := strings.TrimSpace(strings.ToLower(raw))
	spec, ok := hubspotWriteActions[action]
	if !ok {
		return "", writeAction{}, fmt.Errorf("hubspot write action %q is not in the approved allow-list", strings.TrimSpace(raw))
	}
	return action, spec, nil
}

func validateWriteRecord(action string, spec writeAction, record connectors.Record) error {
	if record == nil {
		return fmt.Errorf("%s requires a record", action)
	}
	if spec.needsID && stringField(record, "id") == "" {
		return fmt.Errorf("%s requires an id field", action)
	}
	if action == "create_contact" && stringField(record, "email") == "" {
		return fmt.Errorf("%s requires an email", action)
	}
	if len(contactProperties(record)) == 0 {
		return fmt.Errorf("%s requires at least one writable property", action)
	}
	return nil
}

// contactProperties builds the HubSpot properties map from the allow-listed
// mutable contact fields. The id is carried in the path, not the body, so it is
// excluded here; only non-empty allow-listed properties are included.
func contactProperties(record connectors.Record) map[string]any {
	props := map[string]any{}
	for _, field := range contactWritableProperties {
		if value := stringField(record, field); value != "" {
			props[field] = value
		}
	}
	return props
}
