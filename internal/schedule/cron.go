package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronExpr is a parsed 5-field cron expression.
type CronExpr struct {
	raw     string
	minutes []int // 0-59
	hours   []int // 0-23
	doms    []int // 1-31
	months  []int // 1-12
	dows    []int // 0-6 (0=Sunday)
}

// ParseCron parses and validates a 5-field cron expression.
func ParseCron(expr string) (CronExpr, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return CronExpr{}, fmt.Errorf("cron: expected 5 fields, got %d", len(fields))
	}

	c := CronExpr{raw: expr}
	var err error

	if c.minutes, err = parseField(fields[0], 0, 59); err != nil {
		return CronExpr{}, fmt.Errorf("cron minute: %w", err)
	}
	if c.hours, err = parseField(fields[1], 0, 23); err != nil {
		return CronExpr{}, fmt.Errorf("cron hour: %w", err)
	}
	if c.doms, err = parseField(fields[2], 1, 31); err != nil {
		return CronExpr{}, fmt.Errorf("cron dom: %w", err)
	}
	if c.months, err = parseField(fields[3], 1, 12); err != nil {
		return CronExpr{}, fmt.Errorf("cron month: %w", err)
	}
	if c.dows, err = parseField(fields[4], 0, 6); err != nil {
		return CronExpr{}, fmt.Errorf("cron dow: %w", err)
	}

	return c, nil
}

// parseField parses a single cron field into the set of matching integers.
func parseField(f string, min, max int) ([]int, error) {
	set := map[int]struct{}{}

	// Handle comma-separated list
	parts := strings.Split(f, ",")
	for _, part := range parts {
		vals, err := parsePart(part, min, max)
		if err != nil {
			return nil, err
		}
		for _, v := range vals {
			set[v] = struct{}{}
		}
	}

	result := make([]int, 0, len(set))
	for i := min; i <= max; i++ {
		if _, ok := set[i]; ok {
			result = append(result, i)
		}
	}
	return result, nil
}

func parsePart(part string, min, max int) ([]int, error) {
	// Handle step: */n or a-b/n
	stepStr := ""
	if idx := strings.Index(part, "/"); idx >= 0 {
		stepStr = part[idx+1:]
		part = part[:idx]
	}

	step := 1
	if stepStr != "" {
		n, err := strconv.Atoi(stepStr)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("invalid step %q", stepStr)
		}
		step = n
	}

	var start, end int

	if part == "*" {
		start = min
		end = max
	} else if idx := strings.Index(part, "-"); idx >= 0 {
		// range
		a, err1 := strconv.Atoi(part[:idx])
		b, err2 := strconv.Atoi(part[idx+1:])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("invalid range %q", part)
		}
		if a < min || b > max || a > b {
			return nil, fmt.Errorf("range %d-%d out of bounds [%d,%d]", a, b, min, max)
		}
		start = a
		end = b
	} else {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q", part)
		}
		if n < min || n > max {
			return nil, fmt.Errorf("value %d out of bounds [%d,%d]", n, min, max)
		}
		if stepStr != "" {
			start = n
			end = max
		} else {
			return []int{n}, nil
		}
	}

	var result []int
	for i := start; i <= end; i += step {
		result = append(result, i)
	}
	return result, nil
}

// Next returns the next time the cron expression fires after t (exclusive).
func (c CronExpr) Next(t time.Time) time.Time {
	// Start from t+1 minute (truncated to minute)
	next := t.UTC().Truncate(time.Minute).Add(time.Minute)

	// Search up to 4 years to avoid infinite loop
	limit := next.Add(4 * 365 * 24 * time.Hour)

	for next.Before(limit) {
		// Check month
		if !contains(c.months, int(next.Month())) {
			// Advance to first day of next month
			next = time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			continue
		}
		// Check dom
		if !contains(c.doms, next.Day()) {
			next = time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, time.UTC)
			continue
		}
		// Check dow
		dow := int(next.Weekday()) // 0=Sunday
		if !contains(c.dows, dow) {
			next = time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, time.UTC)
			continue
		}
		// Check hour
		if !contains(c.hours, next.Hour()) {
			nextHour := -1
			for _, h := range c.hours {
				if h > next.Hour() {
					nextHour = h
					break
				}
			}
			if nextHour == -1 {
				// Advance to next day
				next = time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, time.UTC)
			} else {
				next = time.Date(next.Year(), next.Month(), next.Day(), nextHour, 0, 0, 0, time.UTC)
			}
			continue
		}
		// Check minute
		if !contains(c.minutes, next.Minute()) {
			nextMin := -1
			for _, m := range c.minutes {
				if m > next.Minute() {
					nextMin = m
					break
				}
			}
			if nextMin == -1 {
				// Advance to next hour
				next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour()+1, 0, 0, 0, time.UTC)
			} else {
				next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), nextMin, 0, 0, time.UTC)
			}
			continue
		}
		return next
	}
	return time.Time{}
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// Raw returns the original cron expression string.
func (c CronExpr) Raw() string { return c.raw }
