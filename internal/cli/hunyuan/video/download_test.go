package video

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

func TestDownload_MissingArgument(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "download", "-o", "out.mp4")
	if err == nil {
		t.Fatal("expected error for missing argument")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_argument" {
		t.Errorf("expected error code 'missing_argument', got: %s", errorObj["code"])
	}
}

func TestDownload_MissingOutput(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "download", "job-123")
	if err == nil {
		t.Fatal("expected error for missing output")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_output" {
		t.Errorf("expected error code 'missing_output', got: %s", errorObj["code"])
	}
}

func TestDownload_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "download", "job-123", "-o", "out.mp4")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_api_key" {
		t.Errorf("expected error code 'missing_api_key', got: %s", errorObj["code"])
	}
}

func TestDownload_URLSkipsAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "download", "https://invalid.example.com/video.mp4", "-o", "/tmp/test-dl.mp4")
	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] == "missing_api_key" {
		t.Error("URL download should not require API key")
	}
}

func TestDownload_AllFlags(t *testing.T) {
	cmd := NewCmd()
	dlCmd, _, _ := cmd.Find([]string{"download"})
	if dlCmd == nil {
		t.Fatal("download command not found")
	}

	expectedFlags := []string{"output", "region"}
	for _, name := range expectedFlags {
		if dlCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s not found", name)
		}
	}
}

func TestDownload_ShortFlags(t *testing.T) {
	cmd := NewCmd()
	dlCmd, _, _ := cmd.Find([]string{"download"})
	if dlCmd == nil {
		t.Fatal("download command not found")
	}

	flag := dlCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("flag --output not found")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected short flag -o for --output, got -%s", flag.Shorthand)
	}
}
