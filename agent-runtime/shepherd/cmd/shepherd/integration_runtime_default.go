//go:build !integration

package main

import "time"

func integrationGSDExecutor() (string, error) { return "", nil }

func integrationFinalGateBoundary() {}

func integrationEffectTTL() time.Duration { return effectClaimTTL }

func integrationDecisionTTL() time.Duration { return 24 * time.Hour }

func integrationEffectEnqueuedBoundary() {}

func integrationPostValidationBoundary(string) error { return nil }

func integrationRetryBoundary() {}

func integrationExitAwaitingDecision() bool { return false }
