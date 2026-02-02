package video

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/WHQ25/rawgenai/internal/cli/common"
)

// ===== Extend Command Tests =====

func TestExtend_MissingVideoID(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "extend")

	if err == nil {
		t.Fatal("expected error for missing video ID")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "missing_video_id" {
		t.Errorf("expected error code 'missing_video_id', got: %s", errorObj["code"])
	}
}

func TestExtend_MissingAPIKey(t *testing.T) {
	common.SetupNoConfigEnv(t)
	t.Setenv("KLING_ACCESS_KEY", "")
	t.Setenv("KLING_SECRET_KEY", "")

	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "extend", "video-123")

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

func TestExtend_AllFlags(t *testing.T) {
	cmd := newExtendCmd()

	flags := []string{"prompt", "negative", "cfg-scale", "watermark"}
	for _, flag := range flags {
		if cmd.Flag(flag) == nil {
			t.Errorf("expected --%s flag", flag)
		}
	}
}

func TestExtend_DefaultValues(t *testing.T) {
	cmd := newExtendCmd()

	if cmd.Flag("cfg-scale").DefValue != "0.5" {
		t.Errorf("expected default cfg-scale '0.5', got: %s", cmd.Flag("cfg-scale").DefValue)
	}
	if cmd.Flag("watermark").DefValue != "false" {
		t.Errorf("expected default watermark 'false', got: %s", cmd.Flag("watermark").DefValue)
	}
}

func TestExtend_InvalidCfgScale(t *testing.T) {
	cmd := NewCmd()
	_, stderr, err := executeCommand(cmd, "extend", "video-123", "--cfg-scale", "1.5")

	if err == nil {
		t.Fatal("expected error for invalid cfg-scale")
	}

	var resp map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &resp); jsonErr != nil {
		t.Fatalf("expected JSON error output, got: %s", stderr)
	}

	errorObj := resp["error"].(map[string]any)
	if errorObj["code"] != "invalid_cfg_scale" {
		t.Errorf("expected error code 'invalid_cfg_scale', got: %s", errorObj["code"])
	}
}
