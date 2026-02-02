package image

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStatus_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "status")

	// cobra.ExactArgs(1) handles this with Cobra error, not JSON
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	// The error message should mention missing arguments
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("expected error about missing args, got: %s", err.Error())
	}
}

func TestStatus_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "status", "test-task-id")

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestStatus_AllFlags(t *testing.T) {
	cmd := newStatusCmd()

	if cmd.Flags().Lookup("verbose") == nil {
		t.Error("expected flag 'verbose' not found")
	}
}

func TestStatus_ShortFlags(t *testing.T) {
	cmd := newStatusCmd()

	flag := cmd.Flags().ShorthandLookup("v")
	if flag == nil {
		t.Error("expected short flag '-v' not found")
		return
	}
	if flag.Name != "verbose" {
		t.Errorf("short flag '-v' maps to '%s', expected 'verbose'", flag.Name)
	}
}

func TestStatus_DefaultValues(t *testing.T) {
	cmd := newStatusCmd()

	verbose := cmd.Flags().Lookup("verbose")
	if verbose == nil {
		t.Error("flag 'verbose' not found")
		return
	}
	if verbose.DefValue != "false" {
		t.Errorf("flag 'verbose' default is '%s', expected 'false'", verbose.DefValue)
	}
}
