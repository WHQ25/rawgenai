package video

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVideoStatus_MissingTaskID(t *testing.T) {
	cmd := newTestCmd()
	_, _, err := executeCommand(cmd, "status")
	if err == nil {
		t.Fatal("expected error for missing task_id")
	}
}

func TestVideoDownload_MissingOutput(t *testing.T) {
	cmd := newTestCmd()
	_, stderr, err := executeCommand(cmd, "download", "123")
	if err == nil {
		t.Fatal("expected error for missing output")
	}
	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}
}
