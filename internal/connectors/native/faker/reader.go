package faker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// Read generates count deterministic records for stream (default "users"
// when unset, matching legacy's Read, faker.go:47-51). Ported rule-for-rule
// from legacy: same field derivations, same defaults, same error wording
// classification (see docs.md's parity notes).
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	count, err := positiveInt(req.Config.Config["count"], 1000)
	if err != nil {
		return err
	}
	seed, err := parseSeed(req.Config.Config["seed"])
	if err != nil {
		return err
	}

	switch stream {
	case "users":
		return readUsers(ctx, count, seed, emit)
	case "purchases":
		return readPurchases(ctx, count, seed, emit)
	case "products":
		return readProducts(ctx, emit)
	default:
		return fmt.Errorf("faker stream %q not found", stream)
	}
}

// readUsers emits count user records (faker.go:64-73): id/name are
// zero-padded on seed+i, email is user%03d@example.com, updated_at cycles
// through a fixed 28-day window keyed by the loop index alone (not seed) —
// ported verbatim, including that asymmetry, from legacy.
func readUsers(ctx context.Context, count, seed int, emit func(connectors.Record) error) error {
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		n := seed + i
		rec := connectors.Record{
			"id":         fmt.Sprintf("user_%03d", n),
			"name":       fmt.Sprintf("User %03d", n),
			"email":      fmt.Sprintf("user%03d@example.com", n),
			"updated_at": timestamp(i),
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// readPurchases emits count purchase records tied to a generated user and
// one of 10 generated products (faker.go:74-83), ported verbatim.
func readPurchases(ctx context.Context, count, seed int, emit func(connectors.Record) error) error {
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		n := seed + i
		rec := connectors.Record{
			"id":         fmt.Sprintf("purchase_%03d", n),
			"user_id":    fmt.Sprintf("user_%03d", n),
			"product_id": fmt.Sprintf("product_%03d", (i%10)+1),
			"amount":     float64((i%10)+1) * 9.99,
			"updated_at": timestamp(i),
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// readProducts always emits exactly 10 records regardless of count
// (faker.go:84-92), ported verbatim.
func readProducts(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 10; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{
			"id":    fmt.Sprintf("product_%03d", i),
			"sku":   fmt.Sprintf("SKU-%03d", i),
			"name":  fmt.Sprintf("Product %03d", i),
			"price": float64(i) * 4.25,
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// positiveInt parses raw as a positive integer, defaulting to def when raw
// is blank; ported verbatim from legacy's positiveInt (faker.go:103-112).
func positiveInt(raw string, def int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("count must be a positive integer")
	}
	return n, nil
}

// parseSeed parses raw as an integer (default 0 when blank), clamping a
// negative value to 0; ported verbatim from legacy's inline seed parsing in
// Read (faker.go:56-62).
func parseSeed(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}
	seed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("faker config seed must be an integer: %w", err)
	}
	if seed < 0 {
		seed = 0
	}
	return seed, nil
}

// timestamp derives a deterministic RFC3339 timestamp cycling through
// 2026-01-01..2026-01-28 keyed by the loop index i; ported verbatim from
// legacy's timestamp (faker.go:114).
func timestamp(i int) string { return fmt.Sprintf("2026-01-%02dT00:00:00Z", ((i-1)%28)+1) }
