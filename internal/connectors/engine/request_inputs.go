package engine

import (
	"encoding/json"
	"fmt"
)

// RequestInputContract describes generated inputs for one selected consumer.
// It carries executable coordinates only, never authoring source or proof paths.
type RequestInputContract struct {
	Version  int                   `json:"version"`
	Schema   string                `json:"schema"`
	Bindings []RequestInputBinding `json:"bindings"`
}

// RequestInputBinding binds a saved alias to one declared wire coordinate.
type RequestInputBinding struct {
	ConfigKey string `json:"config_key,omitempty"`
	In        string `json:"in"`
	Name      string `json:"name,omitempty"`
	Pointer   string `json:"pointer,omitempty"`
}

type compiledRequestInputPlan struct {
	schema     *Schema
	raw        json.RawMessage
	bodySchema json.RawMessage
	bindings   []RequestInputBinding
}

// streamRequestBodySchema supplies detached schema bytes to body placement.
// A declared contract without its loaded plan cannot use a legacy fallback.
func streamRequestBodySchema(stream StreamSpec) (json.RawMessage, error) {
	if stream.RequestInputs == nil {
		return nil, nil
	}
	if stream.inputPlan == nil {
		return nil, fmt.Errorf("request input contract is not compiled")
	}
	return append(json.RawMessage(nil), stream.inputPlan.bodySchema...), nil
}

// RequestInputSchema returns a detached copy of the schema actually compiled
// for this loaded stream, for canonical admission and definition identity.
func (stream StreamSpec) RequestInputSchema() json.RawMessage {
	if stream.inputPlan == nil {
		return nil
	}
	return append(json.RawMessage(nil), stream.inputPlan.raw...)
}

// RequestInputSchema returns the loaded operation's compiled input schema.
func (operation RESTOperationSpec) RequestInputSchema() json.RawMessage {
	if operation.inputPlan == nil {
		return nil
	}
	return append(json.RawMessage(nil), operation.inputPlan.raw...)
}
