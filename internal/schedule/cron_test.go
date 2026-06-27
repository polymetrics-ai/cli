package schedule

import (
	"testing"
	"time"
)

// Group A — cron parsing and Next() computation.

func TestParseCron_Valid(t *testing.T) {
	cases := []struct {
		id   string
		expr string
	}{
		{"A-1", "0 2 * * *"},
		{"A-2", "*/15 * * * *"},
		{"A-3", "0 9-17 * * 1-5"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.id, func(t *testing.T) {
			_, err := ParseCron(tc.expr)
			if err != nil {
				t.Fatalf("ParseCron(%q) returned unexpected error: %v", tc.expr, err)
			}
		})
	}
}

func TestParseCron_Invalid(t *testing.T) {
	cases := []struct {
		id   string
		expr string
	}{
		{"A-4", "0 2 * *"},     // too few fields
		{"A-5", "0 2 * * * *"}, // too many fields
		{"A-6", "60 * * * *"},  // minute out of range
		{"A-7", "0 25 * * *"},  // hour out of range
		{"A-8", "0 0 1 13 *"},  // month out of range
		{"A-9", "0 0 * * 8"},   // dow out of range
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.id, func(t *testing.T) {
			_, err := ParseCron(tc.expr)
			if err == nil {
				t.Fatalf("ParseCron(%q) expected error but got nil", tc.expr)
			}
		})
	}
}

func TestCronExprNext(t *testing.T) {
	cases := []struct {
		id   string
		expr string
		from time.Time
		want time.Time
	}{
		{
			"A-10",
			"0 2 * * *",
			time.Date(2026, 6, 27, 1, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 27, 2, 0, 0, 0, time.UTC),
		},
		{
			"A-11",
			"0 2 * * *",
			time.Date(2026, 6, 27, 3, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 28, 2, 0, 0, 0, time.UTC),
		},
		{
			"A-12",
			"0 9 * * 1",
			// Monday 2026-06-29 10:00 UTC — already past 09:00 on a Monday, so next Monday
			time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 7, 6, 9, 0, 0, 0, time.UTC),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.id, func(t *testing.T) {
			c, err := ParseCron(tc.expr)
			if err != nil {
				t.Fatalf("ParseCron(%q): %v", tc.expr, err)
			}
			got := c.Next(tc.from)
			if !got.Equal(tc.want) {
				t.Fatalf("Next(%v) = %v, want %v", tc.from, got, tc.want)
			}
		})
	}
}
