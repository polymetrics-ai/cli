package certify

import "testing"

func TestSweeperHarnessRootUsesProjectRootWhenLedgerIsSeparate(t *testing.T) {
	sweeper := NewSweeper(SweeperOptions{Root: "/durable/ledger", ProjectRoot: "/certification/project"})
	if got, want := sweeper.harnessRoot(), "/certification/project"; got != want {
		t.Fatalf("harnessRoot() = %q, want %q", got, want)
	}

	defaultSweeper := NewSweeper(SweeperOptions{Root: "/legacy/project"})
	if got, want := defaultSweeper.harnessRoot(), "/legacy/project"; got != want {
		t.Fatalf("default harnessRoot() = %q, want %q", got, want)
	}
}
