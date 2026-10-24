package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".gac.json")
	t.Setenv("GAC_CONFIG_PATH", configPath)

	// Load non-existent file
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected nil error loading non-existent config, got %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.Active != "" {
		t.Fatalf("expected empty active profile, got %q", cfg.Active)
	}

	// Add profile and save
	cfg.Active = "work"
	cfg.Profiles["work"] = Profile{
		Name:   "Work User",
		Email:  "work@example.com",
		SSHKey: "~/.ssh/id_ed25519_work",
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file mode 0600, got %o", perm)
	}

	// Reload
	reloaded, err := Load()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if reloaded.Active != "work" {
		t.Errorf("expected active 'work', got %q", reloaded.Active)
	}
	profile, ok := reloaded.Profiles["work"]
	if !ok {
		t.Fatalf("expected profile 'work' to exist")
	}
	if profile.Name != "Work User" || profile.Email != "work@example.com" || profile.SSHKey != "~/.ssh/id_ed25519_work" {
		t.Errorf("unexpected profile data: %+v", profile)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"tilde root", "~", home, false},
		{"tilde subpath", "~/.ssh/id_rsa", filepath.Join(home, ".ssh/id_rsa"), false},
		{"plain absolute", "/var/keys/key", "/var/keys/key", false},
		{"empty path", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExpandPath(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}
