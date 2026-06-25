package stripe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics/internal/connectors"
)

// writeAction describes an allow-listed reverse-ETL action: the HTTP method and
// the endpoint path it targets. update_customer appends the customer id to the
// path at execution time.
type writeAction struct {
	method   string
	resource string
	// needsID is true when the action mutates an existing object addressed by id.
	needsID bool
}

// stripeWriteActions is the reverse-ETL allow-list. Anything not present here is
// rejected by ValidateWrite. Starts minimal (customer create/update) to prove
// the write template; expand by adding entries.
var stripeWriteActions = map[string]writeAction{
	"create_customer": {method: http.MethodPost, resource: "customers"},
	"update_customer": {method: http.MethodPost, resource: "customers", needsID: true},
}

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
			return fmt.Errorf("stripe %s record %d: %w", action, i+1, err)
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
		Warnings:      []string{"stripe write executes a live mutation only after approval; dry run performs no external call"},
	}, nil
}

// Write executes an allow-listed action by POSTing form-encoded payloads to the
// Stripe endpoint. Secrets and record PII are never logged; errors reference the
// action and record index only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	action, spec, _ := resolveWriteAction(req.Action)

	if fixtureMode(req.Config) {
		// Fixture mode never mutates the live account; report a receipt-style
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
		path := spec.resource
		if spec.needsID {
			id := stringField(record, "id")
			if id == "" {
				result.RecordsFailed = len(records) - result.RecordsWritten
				return result, fmt.Errorf("stripe %s record %d: missing id", action, i+1)
			}
			path = spec.resource + "/" + url.PathEscape(id)
		}
		form := customerForm(record)
		if _, err := r.DoForm(ctx, spec.method, path, nil, form); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("stripe %s record %d: %w", action, i+1, err)
		}
		result.RecordsWritten++
	}
	return result, nil
}

// resolveWriteAction normalizes and looks up an action against the allow-list.
func resolveWriteAction(raw string) (string, writeAction, error) {
	action := strings.TrimSpace(strings.ToLower(raw))
	spec, ok := stripeWriteActions[action]
	if !ok {
		return "", writeAction{}, fmt.Errorf("stripe write action %q is not in the approved allow-list", strings.TrimSpace(raw))
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
	if action == "create_customer" && stringField(record, "email") == "" && stringField(record, "name") == "" {
		return fmt.Errorf("%s requires at least an email or name", action)
	}
	return nil
}

// customerForm builds the form-encoded body for customer create/update from the
// mutable subset of fields Stripe accepts. The id is carried in the path, not
// the body, so it is excluded here.
func customerForm(record connectors.Record) url.Values {
	form := url.Values{}
	for _, field := range []string{"email", "name", "description", "phone"} {
		if value := stringField(record, field); value != "" {
			form.Set(field, value)
		}
	}
	return form
}
