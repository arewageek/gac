package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIFlow(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".gac.json")
	t.Setenv("GAC_CONFIG_PATH", configPath)

	// Step 1: gac ls on empty config
	var stdout, stderr bytes.Buffer
	err := run([]string{"ls"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run ls failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "No profiles configured") {
		t.Fatalf("expected 'No profiles configured', got: %s", stdout.String())
	}

	// Step 2: gac add work
	fakeKeyPath := filepath.Join(tempDir, "id_ed25519_work")
	if err := os.WriteFile(fakeKeyPath, []byte("fake-ssh-key"), 0600); err != nil {
		t.Fatalf("failed to create fake key: %v", err)
	}

	input := "Arewa Work\nwork@arewageek.com\n" + fakeKeyPath + "\n"
	stdout.Reset()
	stderr.Reset()
	err = run([]string{"add", "work"}, strings.NewReader(input), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run add failed: %v, stderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Saved profile \"work\"") {
		t.Fatalf("unexpected add output: %s", stdout.String())
	}

	// Step 3: gac add personal
	fakePersonalKey := filepath.Join(tempDir, "id_ed25519_personal")
	if err := os.WriteFile(fakePersonalKey, []byte("fake-ssh-key-2"), 0600); err != nil {
		t.Fatalf("failed to create fake personal key: %v", err)
	}

	input2 := "Arewa Geek\nhello@arewageek.com\n" + fakePersonalKey + "\n"
	stdout.Reset()
	stderr.Reset()
	err = run([]string{"add", "personal"}, strings.NewReader(input2), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run add 2 failed: %v", err)
	}

	// Step 4: gac ls
	stdout.Reset()
	err = run([]string{"ls"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run ls failed: %v", err)
	}
	outStr := stdout.String()
	if !strings.Contains(outStr, "work") || !strings.Contains(outStr, "personal") {
		t.Fatalf("expected both profiles in ls output, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "*") || !strings.Contains(outStr, "work") {
		t.Fatalf("expected work to be active with *, got:\n%s", outStr)
	}

	// Step 5: Test gac use --local inside a git repo
	gitRepoDir := filepath.Join(tempDir, "myrepo")
	if out, err := exec.Command("git", "init", gitRepoDir).CombinedOutput(); err != nil {
		t.Fatalf("failed to git init: %s (%v)", string(out), err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(gitRepoDir); err != nil {
		t.Fatalf("failed to chdir to git repo: %v", err)
	}

	stdout.Reset()
	err = run([]string{"use", "personal", "--local"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run use --local failed: %v", err)
	}

	// Verify local git repo config values
	getName, _ := exec.Command("git", "config", "--local", "user.name").Output()
	if strings.TrimSpace(string(getName)) != "Arewa Geek" {
		t.Errorf("expected local user.name 'Arewa Geek', got %q", strings.TrimSpace(string(getName)))
	}

	getEmail, _ := exec.Command("git", "config", "--local", "user.email").Output()
	if strings.TrimSpace(string(getEmail)) != "hello@arewageek.com" {
		t.Errorf("expected local user.email 'hello@arewageek.com', got %q", strings.TrimSpace(string(getEmail)))
	}

	getSSH, _ := exec.Command("git", "config", "--local", "core.sshCommand").Output()
	expectedSSH := "ssh -i " + fakePersonalKey + " -F /dev/null"
	if strings.TrimSpace(string(getSSH)) != expectedSSH {
		t.Errorf("expected local core.sshCommand %q, got %q", expectedSSH, strings.TrimSpace(string(getSSH)))
	}

	// Step 6: gac rm
	stdout.Reset()
	err = run([]string{"rm", "personal"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run rm failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "Removed profile \"personal\"") {
		t.Fatalf("unexpected rm output: %s", stdout.String())
	}
}
