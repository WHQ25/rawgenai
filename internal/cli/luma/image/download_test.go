package image

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDownload_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "download")

	// cobra.ExactArgs(1) handles this with Cobra error, not JSON
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	// The error message should mention missing arguments
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("expected error about missing args, got: %s", err.Error())
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id")

	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestDownload_InvalidOutput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id", "-o", "output.mp4")

	if err == nil {
		t.Fatal("expected error for invalid output extension")
	}

	var resp map[string]any
	json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp)

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_output" {
		t.Errorf("expected error code 'invalid_output', got: %s", errorObj["code"])
	}
}

func TestDownload_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "test-task-id", "-o", "output.jpg")

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

func TestDownload_AllFlags(t *testing.T) {
	cmd := newDownloadCmd()

	if cmd.Flags().Lookup("output") == nil {
		t.Error("expected flag 'output' not found")
	}
}

func TestDownload_ShortFlags(t *testing.T) {
	cmd := newDownloadCmd()

	flag := cmd.Flags().ShorthandLookup("o")
	if flag == nil {
		t.Error("expected short flag '-o' not found")
		return
	}
	if flag.Name != "output" {
		t.Errorf("short flag '-o' maps to '%s', expected 'output'", flag.Name)
	}
}
