package image

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

func TestStatus_MissingJobID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "status")
	if err == nil {
		t.Fatal("expected error for missing job ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_job_id" {
		t.Errorf("expected error code 'missing_job_id', got: %s", errorObj["code"])
	}
}

func TestStatus_EmptyJobID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "status", "  ")
	if err == nil {
		t.Fatal("expected error for empty job ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_job_id" {
		t.Errorf("expected error code 'missing_job_id', got: %s", errorObj["code"])
	}
}

func TestStatus_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("TENCENT_SECRET_ID", "")
	t.Setenv("TENCENT_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "status", "job-123")
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

func TestStatus_AllFlags(t *testing.T) {
	cmd := NewCmd()
	statusCmd, _, _ := cmd.Find([]string{"status"})
	if statusCmd == nil {
		t.Fatal("status command not found")
	}

	expectedFlags := []string{"verbose", "region"}
	for _, name := range expectedFlags {
		if statusCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s not found", name)
		}
	}
}

func TestStatus_ShortFlags(t *testing.T) {
	cmd := NewCmd()
	statusCmd, _, _ := cmd.Find([]string{"status"})
	if statusCmd == nil {
		t.Fatal("status command not found")
	}

	flag := statusCmd.Flags().Lookup("verbose")
	if flag == nil {
		t.Fatal("flag --verbose not found")
	}
	if flag.Shorthand != "v" {
		t.Errorf("expected short flag -v for --verbose, got -%s", flag.Shorthand)
	}
}

func TestStatus_DefaultValues(t *testing.T) {
	cmd := NewCmd()
	statusCmd, _, _ := cmd.Find([]string{"status"})
	if statusCmd == nil {
		t.Fatal("status command not found")
	}

	defaults := map[string]string{
		"verbose": "false",
		"region":  "ap-guangzhou",
	}

	for name, expected := range defaults {
		flag := statusCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("flag --%s not found", name)
			continue
		}
		if flag.DefValue != expected {
			t.Errorf("expected default %s for --%s, got %s", expected, name, flag.DefValue)
		}
	}
}
