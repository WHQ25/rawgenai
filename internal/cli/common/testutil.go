package common

import (
	"os"
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
