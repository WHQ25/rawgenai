package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// SetupNoConfigEnv sets HOME to a temp directory so config file won't be found.
// Call this in tests that check for missing API key errors.
func SetupNoConfigEnv(t *testing.T) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test_no_config")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
}

// SetupConfigWithAPIKey creates a temp config file with the given key-value pairs.
// This simulates having API keys in ~/.config/rawgenai/config.json without env vars.
func SetupConfigWithAPIKey(t *testing.T, keys map[string]string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "test_config")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	configDir := filepath.Join(tmpDir, ".config", "rawgenai")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(keys)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
