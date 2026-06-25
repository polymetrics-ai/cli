package churnkey

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Churnkey Data API resource path
// (relative to base_url), whether it supports limit/skip offset pagination, and
// the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Churnkey Data API path segment (e.g. "sessions"). Churnkey
	// uses a hyphenated path for aggregation ("session-aggregation").
	resource string
	// paginated is true for resources that accept limit/skip offset pagination.
	// session-aggregation returns a bounded, unpaginated array.
	paginated bool
	// mapRecord flattens a raw Churnkey object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// churnkeyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in churnkeyStreams; the read
// path is fully data-driven from this table.
var churnkeyStreamEndpoints = map[string]streamEndpoint{
	"sessions":            {resource: "sessions", paginated: true, mapRecord: churnkeySessionRecord},
	"session_aggregation": {resource: "session-aggregation", paginated: false, mapRecord: churnkeyAggregationRecord},
}

// churnkeyStreams returns the connector's published stream catalog. Sessions are
// keyed by their Mongo-style "_id" and carry a "createdAt" timestamp used as the
// incremental cursor. session_aggregation is a derived rollup with no stable id.
func churnkeyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "sessions",
			Description:  "Churnkey cancel-flow sessions: each customer's pass through a cancellation flow, including the offer accepted, survey responses, and whether they canceled.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"created_at"},
			Fields:       churnkeySessionFields(),
		},
		{
			Name:        "session_aggregation",
			Description: "Aggregated session counts grouped by dimensions such as month, offer type, billing interval, and save type.",
			Fields:      churnkeyAggregationFields(),
		},
	}
}

func churnkeySessionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "org", Type: "string"},
		{Name: "blueprint_id", Type: "string"},
		{Name: "segment_id", Type: "string"},
		{Name: "abtest", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "customer_email", Type: "string"},
		{Name: "customer_plan_id", Type: "string"},
		{Name: "customer_billing_interval", Type: "string"},
		{Name: "offer_type", Type: "string"},
		{Name: "provider", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "aborted", Type: "boolean"},
		{Name: "canceled", Type: "boolean"},
		{Name: "survey_id", Type: "string"},
		{Name: "survey_choice_id", Type: "string"},
		{Name: "survey_choice_value", Type: "string"},
		{Name: "feedback", Type: "string"},
		{Name: "discount_cooldown_applied", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "customer", Type: "object"},
		{Name: "accepted_offer", Type: "object"},
	}
}

func churnkeyAggregationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "count", Type: "integer"},
		{Name: "month", Type: "string"},
		{Name: "trial", Type: "boolean"},
		{Name: "billing_interval", Type: "string"},
		{Name: "plan_id", Type: "string"},
		{Name: "aborted", Type: "boolean"},
		{Name: "canceled", Type: "boolean"},
		{Name: "offer_type", Type: "string"},
		{Name: "save_type", Type: "string"},
	}
}

// churnkeySessionRecord flattens a Churnkey session object. Nested customer and
// acceptedOffer sub-objects are both hoisted into flat columns (for easy
// warehousing) and preserved whole under their original keys.
func churnkeySessionRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"_id":                       item["_id"],
		"org":                       item["org"],
		"blueprint_id":              item["blueprintId"],
		"segment_id":                item["segmentId"],
		"abtest":                    item["abtest"],
		"provider":                  item["provider"],
		"mode":                      item["mode"],
		"aborted":                   item["aborted"],
		"canceled":                  item["canceled"],
		"survey_id":                 item["surveyId"],
		"survey_choice_id":          item["surveyChoiceId"],
		"survey_choice_value":       item["surveyChoiceValue"],
		"feedback":                  item["feedback"],
		"discount_cooldown_applied": item["discountCooldownApplied"],
		"created_at":                item["createdAt"],
		"updated_at":                item["updatedAt"],
		"customer":                  item["customer"],
		"accepted_offer":            item["acceptedOffer"],
	}
	if customer, ok := item["customer"].(map[string]any); ok {
		rec["customer_id"] = customer["id"]
		rec["customer_email"] = customer["email"]
		rec["customer_plan_id"] = customer["planId"]
		rec["customer_billing_interval"] = customer["billingInterval"]
	}
	if offer, ok := item["acceptedOffer"].(map[string]any); ok {
		rec["offer_type"] = offer["offerType"]
	}
	return rec
}

// churnkeyAggregationRecord passes through the grouped count plus whatever
// breakdown dimensions Churnkey returned. The set of dimension keys is
// request-dependent, so the raw object is preserved and the common keys are
// normalized to snake_case.
func churnkeyAggregationRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"count":            item["count"],
		"month":            item["month"],
		"trial":            item["trial"],
		"billing_interval": firstNonNil(item["billingInterval"], item["billing_interval"]),
		"plan_id":          firstNonNil(item["planId"], item["plan_id"]),
		"aborted":          item["aborted"],
		"canceled":         item["canceled"],
		"offer_type":       firstNonNil(item["offerType"], item["offer_type"]),
		"save_type":        firstNonNil(item["saveType"], item["save_type"]),
	}
	// Carry through any additional breakdown dimensions verbatim so nothing is
	// silently dropped.
	for k, v := range item {
		if _, exists := rec[k]; exists {
			continue
		}
		switch k {
		case "billingInterval", "planId", "offerType", "saveType":
			continue // already normalized above
		}
		rec[k] = v
	}
	return rec
}

func firstNonNil(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}
