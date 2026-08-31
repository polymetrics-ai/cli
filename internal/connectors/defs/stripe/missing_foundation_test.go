package stripe

import (
	"strings"
	"testing"
)

const stripeMissingFoundationPath = "missing-foundation.json"

// TestStripeMissingFoundationLedgerStaysSourceBound keeps the one documented
// Stripe webhook registration gap visible without turning a provider endpoint
// configuration into a receiver, executor, or runnable sync transport.
func TestStripeMissingFoundationLedgerStaysSourceBound(t *testing.T) {
	report := readStripeJSONObject(t, stripeMissingFoundationPath)
	if stripeNumberField(report, "schema_version") != 1 || stripeStringField(t, report, "connector") != "stripe" {
		t.Fatalf("Stripe missing-foundation identity = %#v", report)
	}

	lock := readStripeJSONObject(t, stripeSourceLockPath)
	reference := stripeObjectField(t, report, "source_lock")
	lockedREST := stripeObjectField(t, lock, "rest")
	if stripeStringField(t, reference, "path") != stripeSourceLockPath ||
		stripeStringField(t, reference, "source_url") != stripeStringField(t, lockedREST, "source_url") ||
		stripeStringField(t, reference, "sha256") != stripeStringField(t, lockedREST, "sha256") {
		t.Fatalf("Stripe missing-foundation source lock = %#v, want current pinned Stripe lock", reference)
	}

	foundations := stripeArrayField(t, report, "foundations")
	if len(foundations) != 1 {
		t.Fatalf("Stripe missing-foundation entries = %d, want exactly one inbound webhook gap", len(foundations))
	}
	foundation := stripeObjectValue(t, foundations[0], "Stripe missing-foundation entry")
	if stripeStringField(t, foundation, "id") != "stripe-inbound-webhook-receiver-r1" ||
		stripeStringField(t, foundation, "state") != "recorded_only_requires_captain_approval_before_implementation" ||
		stripeStringField(t, foundation, "atlas_capability") != "stripe.inbound-webhook-receiver.v1" ||
		stripeStringField(t, foundation, "affected_lane") != "sync_transport" {
		t.Fatalf("Stripe inbound webhook gap = %#v, want the exact planned receiver record", foundation)
	}

	sourceIDs := stripeArrayField(t, foundation, "source_ids")
	if len(sourceIDs) != 1 {
		t.Fatalf("Stripe inbound webhook source IDs = %#v, want exactly one", sourceIDs)
	}
	source := stripeObjectValue(t, sourceIDs[0], "Stripe inbound webhook source ID")
	if stripeStringField(t, source, "id") != "stripe.rest.PostWebhookEndpoints" ||
		stripeStringField(t, source, "operation_id") != "PostWebhookEndpoints" ||
		stripeStringField(t, source, "method") != "POST" ||
		stripeStringField(t, source, "path") != "/v1/webhook_endpoints" ||
		stripeStringField(t, source, "source_location") != "paths[\"/v1/webhook_endpoints\"].post" {
		t.Fatalf("Stripe inbound webhook source = %#v, want the exact locked operation", source)
	}

	matrix := readStripeJSONObject(t, stripeMatrixPath)
	row := stripeMatrixOperationsByID(t, matrix)[stripeStringField(t, source, "id")]
	if row == nil {
		t.Fatalf("Stripe missing-foundation source row %q is absent from matrix", stripeStringField(t, source, "id"))
	}
	cell := stripeMatrixCellByLane(t, row, "sync_transport")
	if stripeStringField(t, cell, "state") != "missing_foundation" ||
		stripeStringField(t, cell, "reason") != "stripe.foundation.inbound_webhook_receiver.v1" {
		t.Fatalf("Stripe inbound webhook matrix cell = %#v, want named missing foundation", cell)
	}

	claim := strings.ToLower(stripeStringField(t, foundation, "runtime_claim"))
	if !strings.Contains(claim, "no inbound stripe webhook receiver") ||
		!strings.Contains(claim, "no inbound stripe webhook receiver, selected source executor, or executable sync transport is claimed") {
		t.Fatalf("Stripe inbound webhook runtime claim = %q, want explicit non-execution boundary", claim)
	}
}
