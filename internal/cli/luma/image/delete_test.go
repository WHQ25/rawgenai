package image

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDelete_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "delete")

	// cobra.ExactArgs(1) handles this with Cobra error, not JSON
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	// The error message should mention missing arguments
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("expected error about missing args, got: %s", err.Error())
	}
}

func TestDelete_MissingAPIKey(t *testing.T) {
	setupNoConfigEnv(t)

	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "delete", "test-task-id")

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
