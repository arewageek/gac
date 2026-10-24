package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyProfile(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize a dummy git repo for testing
	initCmd := exec.Command("git", "init", tempDir)
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to git init temp dir: %s (%v)", string(out), err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir to temp repo: %v", err)
	}

	if !IsInsideWorkTree() {
		t.Fatalf("expected IsInsideWorkTree to be true in temp git repo")
	}

	keyPath := filepath.Join(tempDir, "id_test")
	err = ApplyProfile("Test Runner", "runner@example.com", keyPath, true)
	if err != nil {
		t.Fatalf("ApplyProfile failed: %v", err)
	}

	// Verify local git configs
	checkConfig := func(key string) string {
		out, err := exec.Command("git", "config", "--local", key).Output()
		if err != nil {
			t.Fatalf("failed to read git config %s: %v", key, err)
		}
		return strings.TrimSpace(string(out))
	}

	if val := checkConfig("user.name"); val != "Test Runner" {
		t.Errorf("expected user.name 'Test Runner', got %q", val)
	}
	if val := checkConfig("user.email"); val != "runner@example.com" {
		t.Errorf("expected user.email 'runner@example.com', got %q", val)
	}
	expectedSSH := "ssh -i " + keyPath + " -F /dev/null"
	if val := checkConfig("core.sshCommand"); val != expectedSSH {
		t.Errorf("expected core.sshCommand %q, got %q", expectedSSH, val)
	}
}

func TestLocalOutsideRepo(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir to non-git temp dir: %v", err)
	}

	err = ApplyProfile("Outside", "outside@example.com", "/dummy/key", true)
	if err == nil {
		t.Fatalf("expected error when applying --local outside git repo, got nil")
	}
}
