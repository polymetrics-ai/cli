package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSentrySeerModelsSourceProjectionKeepsExactRouteBinding(t *testing.T) {
	rawLock, err := os.ReadFile(filepath.Join("testdata", "issue4329", "sentry-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read Sentry source lock: %v", err)
	}
	lock, err := parseSourceImportLock(rawLock, "sentry")
	if err != nil {
		t.Fatalf("parse Sentry source lock: %v", err)
	}

	var source *sourceImportRESTOperation
	for index := range lock.Rest.Operations {
		candidate := &lock.Rest.Operations[index]
		if candidate.ID == "sentry.rest.listSeerModels" {
			source = candidate
			break
		}
	}
	if source == nil {
		t.Fatal("Sentry source lock does not contain sentry.rest.listSeerModels")
	}
	if source.Method != http.MethodGet || source.Path != "/api/0/seer/models/" || source.OperationID != "listSeerModels" {
		t.Fatalf("Sentry Seer Models source = %+v, want exact GET /api/0/seer/models/ listSeerModels", source)
	}

	bundle, err := engine.Load(os.DirFS(filepath.Join("..", "..", "internal", "connectors", "defs")), "sentry")
	if err != nil {
		t.Fatalf("load Sentry bundle: %v", err)
	}
	if bundle.CLISurface == nil || bundle.Surface == nil {
		t.Fatalf("Sentry bundle command/surface declarations are missing: cli=%t surface=%t", bundle.CLISurface != nil, bundle.Surface != nil)
	}

	var operation *engine.OperationSpec
	for index := range bundle.Operations {
		candidate := &bundle.Operations[index]
		if candidate.ID == "sentry.seer_models_list" {
			operation = candidate
			break
		}
	}
	if operation == nil {
		t.Fatal("Sentry bundle does not declare sentry.seer_models_list")
	}
	if operation.Route != "sentry_api_v0" || operation.SourceURL != lock.Rest.SourceURL || operation.SourceOperation == nil || operation.SourceOperation.ID != source.ID || operation.SourceOperation.Method != source.Method || operation.SourceOperation.Path != source.Path || operation.REST == nil || operation.REST.Method != source.Method || operation.REST.Path != source.Path {
		t.Fatalf("Sentry operation = %+v, want the exact lock-bound Seer Models GET on sentry_api_v0", operation)
	}

	commandCount := 0
	for _, command := range bundle.CLISurface.Commands {
		if command.Path != "seer list-models" {
			continue
		}
		commandCount++
		if command.Intent != "direct_read" || command.Availability != "implemented" || command.Operation != operation.ID || command.SourceOperation != source.ID || len(command.APISurface) != 1 || command.APISurface[0].Method != source.Method || command.APISurface[0].Path != source.Path {
			t.Fatalf("Sentry command = %+v, want one exact source-bound Seer Models direct read", command)
		}
	}
	if commandCount != 1 {
		t.Fatalf("Sentry %q command count = %d, want one", "seer list-models", commandCount)
	}

	endpointCount := 0
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Method != source.Method || endpoint.Path != source.Path {
			continue
		}
		endpointCount++
		if endpoint.CoveredBy == nil || endpoint.CoveredBy.DirectRead != "seer list-models" {
			t.Fatalf("Sentry Seer Models api surface = %+v, want direct-read coverage by seer list-models", endpoint)
		}
	}
	if endpointCount != 1 {
		t.Fatalf("Sentry %s %s api surface entries = %d, want one", source.Method, source.Path, endpointCount)
	}
}
