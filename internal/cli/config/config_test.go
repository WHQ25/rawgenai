package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err = cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func setupTestEnv(t *testing.T) (cleanup func()) {
	tmpDir, err := os.MkdirTemp("", "config_cli_test")
	if err != nil {
		t.Fatal(err)
	}

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	return func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tmpDir)
	}
}

func TestConfigPath(t *testing.T) {
	stdout, _, err := executeCommand(Cmd, "path")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp successResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	if resp.Path == "" {
		t.Error("expected path to be non-empty")
	}

	if !strings.Contains(resp.Path, "config.json") {
		t.Errorf("expected path to contain config.json, got: %s", resp.Path)
	}
}

func TestConfigList(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	stdout, _, err := executeCommand(Cmd, "list")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp listResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Empty config should return empty keys map (unset keys are hidden)
	if len(resp.Keys) != 0 {
		t.Errorf("expected empty keys map for new config, got %d keys", len(resp.Keys))
	}
}

func TestConfigSet(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	stdout, _, err := executeCommand(Cmd, "set", "openai_api_key", "sk-test123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp successResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	if !strings.Contains(resp.Message, "openai_api_key") {
		t.Errorf("expected message to contain key name, got: %s", resp.Message)
	}
}

func TestConfigSet_EnvVarFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Should accept OPENAI_API_KEY format
	stdout, _, err := executeCommand(Cmd, "set", "OPENAI_API_KEY", "sk-test123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp successResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestConfigSet_InvalidKey(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, stderr, err := executeCommand(Cmd, "set", "invalid_key", "value")

	if err == nil {
		t.Fatal("expected error for invalid key")
	}

	var resp errorResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp.Success {
		t.Error("expected success to be false")
	}

	if resp.Error.Code != "invalid_key" {
		t.Errorf("expected error code 'invalid_key', got: %s", resp.Error.Code)
	}
}

func TestConfigUnset(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// First set a value
	executeCommand(Cmd, "set", "openai_api_key", "sk-test123")

	// Then unset it
	stdout, _, err := executeCommand(Cmd, "unset", "openai_api_key")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp successResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify it's unset (key should not appear in list)
	listStdout, _, _ := executeCommand(Cmd, "list")
	var listResp listResponse
	json.Unmarshal([]byte(strings.TrimSpace(listStdout)), &listResp)

	if _, exists := listResp.Keys["openai_api_key"]; exists {
		t.Errorf("expected key to not exist in list after unset, got: %s", listResp.Keys["openai_api_key"])
	}
}

func TestConfigUnset_InvalidKey(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, stderr, err := executeCommand(Cmd, "unset", "invalid_key")

	if err == nil {
		t.Fatal("expected error for invalid key")
	}

	var resp errorResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	if resp.Error.Code != "invalid_key" {
		t.Errorf("expected error code 'invalid_key', got: %s", resp.Error.Code)
	}
}

func TestConfigSetAndList(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Set a value
	executeCommand(Cmd, "set", "openai_api_key", "sk-test123abc")

	// List and verify masked value
	stdout, _, err := executeCommand(Cmd, "list")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp listResponse
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}

	// Should be masked
	if resp.Keys["openai_api_key"] != "sk-***abc" {
		t.Errorf("expected masked value 'sk-***abc', got: %s", resp.Keys["openai_api_key"])
	}
}

func TestConfigSet_MissingArgs(t *testing.T) {
	_, _, err := executeCommand(Cmd, "set", "openai_api_key")

	if err == nil {
		t.Fatal("expected error for missing value argument")
	}
}

func TestConfigUnset_MissingArgs(t *testing.T) {
	_, _, err := executeCommand(Cmd, "unset")

	if err == nil {
		t.Fatal("expected error for missing key argument")
	}
}
