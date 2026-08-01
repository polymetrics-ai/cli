package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestShopifyAPISurfaceDestructiveDisposition(t *testing.T) {
	surface := loadShopifySurface(t)

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}
	if len(surface.Endpoints) != 1166 {
		t.Fatalf("endpoints = %d, want 1166", len(surface.Endpoints))
	}

	models := map[string]int{}
	risks := map[string]int{}
	destructiveConfirm := 0
	covered, operations := 0, 0
	byPath := map[string]engine.SurfaceEndpoint{}

	for i, ep := range surface.Endpoints {
		if ep.Path != "" {
			byPath[ep.Path] = ep
		}
		if ep.CoveredBy != nil {
			covered++
		}
		if ep.Operation == nil {
			continue
		}
		operations++
		models[ep.Operation.Model]++
		risks[ep.Operation.Risk]++
		if !ep.Operation.BlockedByDefault {
			t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
		}
		if ep.Operation.Model == "destructive_action" {
			if ep.Operation.Risk != "critical" || ep.Operation.Confirm != "destructive" {
				t.Fatalf("endpoint %d destructive operation = %+v, want critical risk and destructive confirmation", i, ep.Operation)
			}
			destructiveConfirm++
		}
		if shopifyStateDestroying(ep) {
			if ep.Operation.Model != "destructive_action" || ep.Operation.Risk != "critical" || ep.Operation.Confirm != "destructive" {
				t.Fatalf("state-destroying endpoint %s %s = %+v, want destructive_action/critical/destructive", ep.Method, ep.Path, ep.Operation)
			}
		}
	}

	if covered != 43 {
		t.Fatalf("covered endpoints = %d, want 43", covered)
	}
	if operations != 1123 {
		t.Fatalf("operation endpoints = %d, want 1123", operations)
	}
	if destructiveConfirm != 157 {
		t.Fatalf("destructive confirmed operations = %d, want 157", destructiveConfirm)
	}
	assertStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl":  495,
		"binary_read":        17,
		"destructive_action": 157,
		"direct_read":        454,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"critical": 157,
		"high":     3,
		"low":      471,
		"medium":   492,
	})

	for _, path := range []string{
		"GraphQL Mutation.customerRequestDataErasure",
		"GraphQL Mutation.delegateAccessTokenDestroy",
		"GraphQL Mutation.fulfillmentOrderReleaseHold",
		"GraphQL Mutation.locationLocalPickupDisable",
		"GraphQL Mutation.privacyFeaturesDisable",
		"GraphQL Mutation.removeFromReturn",
		"GraphQL Mutation.reverseFulfillmentOrderDispose",
		"GraphQL Mutation.shopLocaleDisable",
		"/admin/api/latest/comments/{comment_id}/remove.json",
		"/admin/api/latest/fulfillment_orders/{fulfillment_order_id}/cancel.json",
		"/admin/api/latest/fulfillment_orders/{fulfillment_order_id}/cancellation_request.json",
		"/admin/api/latest/fulfillment_orders/{fulfillment_order_id}/cancellation_request/accept.json",
		"/admin/api/latest/fulfillment_orders/{fulfillment_order_id}/cancellation_request/reject.json",
		"/admin/api/latest/fulfillment_orders/{fulfillment_order_id}/close.json",
		"/admin/api/latest/fulfillment_orders/{fulfillment_order_id}/release_hold.json",
		"/admin/api/latest/fulfillments/{fulfillment_id}/cancel.json",
		"/admin/api/latest/gift_cards/{gift_card_id}/disable.json",
		"/admin/api/latest/orders/{order_id}/cancel.json",
		"/admin/api/latest/orders/{order_id}/close.json",
	} {
		ep, ok := byPath[path]
		if !ok {
			t.Fatalf("expected endpoint %q", path)
		}
		if ep.Operation == nil || ep.Operation.Model != "destructive_action" || ep.Operation.Risk != "critical" || ep.Operation.Confirm != "destructive" {
			t.Fatalf("endpoint %q operation = %+v, want destructive_action/critical/destructive", path, ep.Operation)
		}
	}

	activate := byPath["GraphQL Mutation.discountCodeActivate"].Operation
	if activate == nil || activate.Model != "admin_reverse_etl" || activate.Risk != "medium" || activate.Confirm != "" {
		t.Fatalf("discountCodeActivate operation = %+v, want non-destructive admin_reverse_etl/medium", activate)
	}
	hold := byPath["GraphQL Mutation.fulfillmentOrderHold"].Operation
	if hold == nil || hold.Model != "admin_reverse_etl" || hold.Risk != "medium" || hold.Confirm != "" {
		t.Fatalf("fulfillmentOrderHold operation = %+v, want non-destructive admin_reverse_etl/medium", hold)
	}
}

func TestShopifyDeleteWritesRequireDestructiveConfirmation(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/shopify/writes.json")
	if err != nil {
		t.Fatalf("read shopify writes.json: %v", err)
	}

	var writes struct {
		Actions []struct {
			Name    string `json:"name"`
			Kind    string `json:"kind"`
			Confirm string `json:"confirm"`
		} `json:"actions"`
	}
	if err := json.Unmarshal(raw, &writes); err != nil {
		t.Fatalf("unmarshal shopify writes.json: %v", err)
	}
	if len(writes.Actions) != 42 {
		t.Fatalf("write actions = %d, want 42", len(writes.Actions))
	}
	for _, action := range writes.Actions {
		if action.Kind != "delete" || action.Confirm != "destructive" {
			t.Fatalf("write action %s = kind %q confirm %q, want delete/destructive", action.Name, action.Kind, action.Confirm)
		}
	}
}

func loadShopifySurface(t *testing.T) engine.APISurface {
	t.Helper()
	bundle, err := engine.Load(os.DirFS("../../internal/connectors/defs"), "shopify")
	if err != nil {
		t.Fatalf("load shopify bundle: %v", err)
	}
	if bundle.Surface == nil {
		t.Fatalf("shopify api_surface.json was not loaded")
	}
	return *bundle.Surface
}

func shopifyStateDestroying(ep engine.SurfaceEndpoint) bool {
	if ep.Operation == nil {
		return false
	}
	if ep.Method == "DELETE" {
		return true
	}
	var tokens []string
	if strings.HasPrefix(ep.Path, "GraphQL Mutation.") {
		tokens = shopifyCamelTokens(strings.TrimPrefix(ep.Path, "GraphQL Mutation."))
	} else {
		tokens = shopifyPathTokens(ep.Path)
	}
	for _, token := range tokens {
		switch token {
		case "archive", "cancel", "close", "deactivate", "delete", "destroy", "disable", "discard", "dispose", "erasure", "expire", "release", "remove", "revoke", "uninstall", "void":
			return true
		}
		if strings.HasPrefix(token, "cancel") {
			return true
		}
	}
	return false
}

func shopifyCamelTokens(s string) []string {
	var tokens []string
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			tokens = append(tokens, strings.ToLower(s[start:end]))
		}
		start = -1
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !shopifyASCIIAlphaNum(c) {
			flush(i)
			continue
		}
		if start == -1 {
			start = i
			continue
		}
		prev := s[i-1]
		nextLower := i+1 < len(s) && shopifyASCIILower(s[i+1])
		if shopifyASCIIUpper(c) && (shopifyASCIILower(prev) || shopifyASCIIDigit(prev) || (shopifyASCIIUpper(prev) && nextLower)) {
			flush(i)
			start = i
		}
	}
	flush(len(s))
	return tokens
}

func shopifyASCIIAlphaNum(c byte) bool {
	return shopifyASCIIUpper(c) || shopifyASCIILower(c) || shopifyASCIIDigit(c)
}

func shopifyASCIIUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func shopifyASCIILower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func shopifyASCIIDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func shopifyPathTokens(path string) []string {
	path = strings.TrimSuffix(strings.ToLower(path), ".json")
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	return parts
}

func TestShopifyTokenizationDoesNotTreatActivateAsDeactivate(t *testing.T) {
	if got := shopifyStateDestroying(engine.SurfaceEndpoint{Method: "POST", Path: "GraphQL Mutation.discountCodeActivate", Operation: &engine.SurfaceOperation{}}); got {
		t.Fatalf("discountCodeActivate classified as state-destroying")
	}
	if !reflect.DeepEqual(shopifyCamelTokens("delegateAccessTokenDestroy"), []string{"delegate", "access", "token", "destroy"}) {
		t.Fatalf("unexpected delegateAccessTokenDestroy tokens")
	}
	if !reflect.DeepEqual(shopifyCamelTokens("reverseFulfillmentOrderDispose"), []string{"reverse", "fulfillment", "order", "dispose"}) {
		t.Fatalf("unexpected reverseFulfillmentOrderDispose tokens")
	}
	if !shopifyStateDestroying(engine.SurfaceEndpoint{Method: "POST", Path: "GraphQL Mutation.fulfillmentOrderReleaseHold", Operation: &engine.SurfaceOperation{}}) {
		t.Fatalf("fulfillmentOrderReleaseHold was not classified as state-destroying")
	}
	if got := shopifyStateDestroying(engine.SurfaceEndpoint{Method: "POST", Path: "GraphQL Mutation.fulfillmentOrderHold", Operation: &engine.SurfaceOperation{}}); got {
		t.Fatalf("fulfillmentOrderHold classified as state-destroying")
	}
}
