package workspace

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRepositoryLockReleasedAfterHardKilledController(t *testing.T) {
	if os.Getenv("SHEPHERD_LOCK_CRASH_HELPER") != "" {
		return
	}
	repo := initRepo(t)
	root := filepath.Join(t.TempDir(), "attempts")
	cmd := exec.Command(os.Args[0], "-test.run=TestRepositoryLockCrashHelper", "--", repo, root)
	cmd.Env = append(os.Environ(), "SHEPHERD_LOCK_CRASH_HELPER=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "locked" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("crash helper did not acquire lock: %v", scanner.Err())
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("hard-killed helper exited successfully")
	}
	manager, err := NewManager(repo, root)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := manager.TryAcquireRepositoryLock()
	if err != nil {
		t.Fatalf("kernel did not release crashed controller lock: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryLockCrashHelper(t *testing.T) {
	if os.Getenv("SHEPHERD_LOCK_CRASH_HELPER") != "1" {
		return
	}
	args := os.Args
	separator := -1
	for i, arg := range args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator < 0 || separator+2 >= len(args) {
		os.Exit(2)
	}
	manager, err := NewManager(args[separator+1], args[separator+2])
	if err != nil {
		os.Exit(2)
	}
	lock, err := manager.TryAcquireRepositoryLock()
	if err != nil {
		os.Exit(2)
	}
	defer func() { _ = lock.Close() }()
	fmt.Println("locked")
	time.Sleep(30 * time.Second)
}

func TestRepositoryLockExcludesConcurrentControllerAndReleases(t *testing.T) {
	repo := initRepo(t)
	firstRoot := filepath.Join(t.TempDir(), "attempts")
	firstManager, err := NewManager(repo, firstRoot)
	if err != nil {
		t.Fatal(err)
	}
	secondManager, err := NewManager(repo, filepath.Join(t.TempDir(), "other-state-attempts"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := firstManager.TryAcquireRepositoryLock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := secondManager.TryAcquireRepositoryLock(); err == nil {
		t.Fatal("concurrent controller acquired repository lock")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := secondManager.TryAcquireRepositoryLock()
	if err != nil {
		t.Fatalf("released repository lock not reusable: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}
