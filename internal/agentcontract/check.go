// Package agentcontract validates and projects the canonical agent delivery contract.
package agentcontract

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CheckGSDCommands verifies command names through the repository's real GSD adapter.
func CheckGSDCommands(ctx context.Context, script string, commands []string) error {
	for _, command := range commands {
		output, err := exec.CommandContext(ctx, script, "sources", command).CombinedOutput()
		if err != nil {
			return fmt.Errorf("GSD command %q does not resolve: %s: %w", command, strings.TrimSpace(string(output)), err)
		}
	}
	return nil
}

// CheckProjection is the pre-enforcement placeholder exercised by the RED test.
func CheckProjection(_, _ []byte) error {
	return nil
}
