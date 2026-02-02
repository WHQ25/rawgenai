package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDelete_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "delete")

	if err == nil {
		t.Fatal("expected error for missing task_id")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_task_id" {
		t.Errorf("expected error code 'missing_task_id', got: %s", errorObj["code"])
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
