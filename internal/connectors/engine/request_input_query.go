package engine

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"polymetrics.ai/internal/connectors"
)

func selectedStructuredQuery(plan *compiledRequestInputPlan, name string) bool {
	if plan == nil || plan.queryEncoding == nil || plan.schema == nil {
		return false
	}
	node := plan.schema.node.properties["query"].properties[name]
	return node != nil && len(node.types) == 1 && (node.types[0] == "object" || node.types[0] == "array")
}

func decodeSelectedRequestInput(plan *compiledRequestInputPlan, binding RequestInputBinding, kind, raw string) (any, error) {
	if binding.In == "query" && selectedStructuredQuery(plan, binding.Name) {
		return connectors.DecodeSourceStructured(kind, raw, 1<<20)
	}
	return connectors.DecodeSourceScalar(kind, raw, 1<<20)
}

func mergePreparedQuery(stream StreamSpec, query url.Values) url.Values {
	if stream.inputPlan == nil || stream.inputPlan.queryEncoding == nil {
		return query
	}
	for name := range stream.inputPlan.schema.node.properties["query"].properties {
		delete(query, name)
	}
	for name, values := range stream.preparedQueryPairs {
		query[name] = append([]string(nil), values...)
	}
	return query
}

// Each encoded coordinate remains owned by its selected source input. A later
// pagination URL cannot replace it, including through a repeated value.
func validatePreparedStructuredQuery(stream StreamSpec, query url.Values) (map[string]bool, error) {
	known := map[string]bool{}
	if stream.inputPlan == nil || stream.inputPlan.queryEncoding == nil {
		return known, nil
	}
	for name, values := range stream.preparedQueryPairs {
		// Scalar coordinates remain subject to effective paging validation.
		if _, scalar := stream.inputPlan.schema.node.properties["query"].properties[name]; scalar && !selectedStructuredQuery(stream.inputPlan, name) {
			continue
		}
		if !reflect.DeepEqual(query[name], values) {
			return nil, fmt.Errorf("effective query changed a structured input")
		}
		known[name] = true
	}
	return known, nil
}

func prepareOperationStructuredValues(plan *compiledRequestInputPlan, values map[string]any) (map[string]any, error) {
	if plan == nil || plan.queryEncoding == nil {
		return nil, fmt.Errorf("structured values require query encoding")
	}
	for name := range values {
		if !selectedStructuredQuery(plan, name) {
			return nil, fmt.Errorf("undeclared structured query field")
		}
	}
	state := formEncodingState{spec: plan.queryEncoding}
	normalized, err := state.normalize(reflect.ValueOf(values), 0)
	if err != nil {
		return nil, err
	}
	return normalized.(map[string]any), nil
}

func operationTypedQuery(op OperationSpec, scalar map[string]string, structured map[string]any) (url.Values, error) {
	plan := op.REST.inputPlan
	if plan == nil || plan.queryEncoding == nil {
		if len(structured) != 0 {
			return nil, fmt.Errorf("structured query lacks its selected encoding")
		}
		return operationDirectReadQueryValues(op, scalar)
	}
	values, err := prepareOperationStructuredValues(plan, structured)
	if err != nil {
		return nil, err
	}
	for name, raw := range scalar {
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("conflicting query value representations")
		}
		node := plan.schema.node.properties["query"].properties[name]
		if node == nil || len(node.types) != 1 || selectedStructuredQuery(plan, name) {
			return nil, fmt.Errorf("undeclared scalar query value")
		}
		value, err := connectors.DecodeSourceScalar(node.types[0], raw, 1<<20)
		if err != nil {
			return nil, err
		}
		values[name] = value
	}
	query, _, err := encodeTypedForm(plan.queryEncoding, plan.querySchema, values, nil)
	return query, err
}

func (c *Connector) PreflightOperationStructuredQueryField(operation, field string) error {
	op, err := findOperation(c.bundle, operation)
	if err != nil {
		return err
	}
	if op.Kind != "rest_read" || op.REST == nil || !selectedStructuredQuery(op.REST.inputPlan, field) {
		return fmt.Errorf("query field lacks selected structured input contract")
	}
	return nil
}

func validateOperationQueryInputs(op *OperationSpec, initial url.Values, structured map[string]any, effective url.Values) error {
	if op == nil || op.REST == nil || op.REST.inputPlan == nil || op.REST.inputPlan.queryEncoding == nil {
		return nil
	}
	plan := op.REST.inputPlan
	known, err := validatePreparedStructuredQuery(StreamSpec{inputPlan: plan, preparedQueryPairs: initial}, effective)
	if err != nil {
		return err
	}
	values := map[string]any{}
	for name, value := range structured {
		values[name] = value
	}
	for name, entries := range effective {
		if known[name] {
			continue
		}
		node := plan.schema.node.properties["query"].properties[name]
		if node != nil {
			if len(entries) != 1 || len(node.types) != 1 {
				return fmt.Errorf("effective scalar query requires one value")
			}
			value, err := connectors.DecodeSourceScalar(node.types[0], entries[0], 1<<20)
			if err != nil {
				return err
			}
			values[name] = value
			continue
		}
		declared := false
		for _, parameter := range op.REST.PaginationParameters {
			if parameter.Name != name {
				continue
			}
			declared = true
			if len(entries) != 1 {
				return fmt.Errorf("effective navigation query requires one value")
			}
			if err := validateOperationParameterWireValue(*op, parameter, "query", entries[0]); err != nil {
				return err
			}
		}
		if !declared {
			return fmt.Errorf("effective query contains undeclared input")
		}
	}
	if err := plan.schema.node.properties["query"].validate(values, ""); err != nil {
		return fmt.Errorf("effective query violates selected schema")
	}
	bounds := formEncodingState{spec: plan.queryEncoding, pairs: url.Values{}}
	for name, entries := range effective {
		for _, entry := range entries {
			if err := bounds.add(name, entry); err != nil {
				return err
			}
		}
	}
	return nil
}

func structuredInputText(plan *compiledRequestInputPlan, value any) (string, error) {
	state := formEncodingState{spec: plan.queryEncoding}
	normalized, err := state.normalize(reflect.ValueOf(value), 0)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("invalid structured input default")
	}
	return string(raw), nil
}
