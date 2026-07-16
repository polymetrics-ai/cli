//go:build integration && !race

package main

import "time"

func integrationShortEffectTTL() time.Duration { return 500 * time.Millisecond }
