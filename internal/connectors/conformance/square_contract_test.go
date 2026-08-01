package conformance

import "testing"

func TestSquarePaymentsSchemaProjectsOpenAPIFields(t *testing.T) {
	b := loadTestBundle(t, "../defs", "square")
	sch := b.Schemas["payments"]
	if sch == nil {
		t.Fatalf("payments schema missing")
	}

	props := propertySet(sch.Properties())
	for _, field := range []string{
		"amount_money",
		"total_money",
		"processing_fee",
		"source_type",
		"location_id",
		"order_id",
		"receipt_number",
	} {
		if !props[field] {
			t.Fatalf("payments schema missing projected field %q", field)
		}
	}

	replay := newReusableStreamReplayServer()
	defer replay.Close()

	var first map[string]any
	if err := readRawRecordsWithReplay(b, "payments", nil, replay, func(record map[string]any) error {
		if first == nil {
			first = record
		}
		return nil
	}); err != nil {
		t.Fatalf("read payments fixture: %v", err)
	}
	if first == nil {
		t.Fatalf("payments fixture emitted no records")
	}
	for _, field := range []string{"amount_money", "total_money", "processing_fee"} {
		if _, ok := first[field]; !ok {
			t.Fatalf("projected payments fixture missing %q: %+v", field, first)
		}
	}
}

func TestSquareStreamSchemasAreNotGenericPlaceholders(t *testing.T) {
	b := loadTestBundle(t, "../defs", "square")
	placeholder := map[string]bool{
		"id":               true,
		"created_at":       true,
		"updated_at":       true,
		"name":             true,
		"status":           true,
		"source_operation": true,
	}
	for name, schema := range b.Schemas {
		props := schema.Properties()
		if len(props) != len(placeholder) {
			continue
		}
		allPlaceholder := true
		for _, prop := range props {
			if !placeholder[prop] {
				allPlaceholder = false
				break
			}
		}
		if allPlaceholder {
			t.Fatalf("schema %q still has generic placeholder projection", name)
		}
	}
}
