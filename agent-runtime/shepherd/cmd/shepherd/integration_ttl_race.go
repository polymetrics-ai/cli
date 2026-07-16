//go:build integration && race

package main

import "time"

func integrationShortEffectTTL() time.Duration { return 15 * time.Second }
